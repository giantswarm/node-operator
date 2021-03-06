package drainer

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/node-operator/service/controller/key"
)

const (
	// UnschedulablePatch is the JSON patch structure being applied to nodes using
	// a strategic merge patch in order to drain them.
	UnschedulablePatch = `{"spec":{"unschedulable":true}}`
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	drainerConfig, err := key.ToDrainerConfig(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if drainerConfig.Status.HasDrainedCondition() {
		r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config status has drained condition")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	if drainerConfig.Status.HasTimeoutCondition() {
		r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config status has timeout condition")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	if drainingTimedOut(drainerConfig, time.Now(), 10*time.Minute) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config exists for too long without draining being finished")
		r.logger.LogCtx(ctx, "level", "debug", "message", "setting drainer config status of tenant cluster node to timeout condition")

		drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewTimeoutCondition())

		_, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(drainerConfig.GetNamespace()).UpdateStatus(ctx, &drainerConfig, metav1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "set drainer config status of tenant cluster node to timeout condition")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	var restConfig *rest.Config
	{
		i := key.ClusterIDFromDrainerConfig(drainerConfig)
		e := key.ClusterEndpointFromDrainerConfig(drainerConfig)
		restConfig, err = r.tenantCluster.NewRestConfig(ctx, i, e)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "fetching certificates timed out")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	var k8sClient kubernetes.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger:     r.logger,
			RestConfig: restConfig,
		}

		k8sClients, err := k8sclient.NewClients(c)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster API is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		k8sClient = k8sClients.K8sClient()
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoning tenant cluster node")

		n := key.NodeNameFromDrainerConfig(drainerConfig)
		t := types.StrategicMergePatchType
		p := []byte(UnschedulablePatch)

		_, err := k8sClient.CoreV1().Nodes().Patch(ctx, n, t, p, metav1.PatchOptions{})
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster API is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if apierrors.IsNotFound(err) {
			// It might happen the node we want to drain got already removed. This
			// might even be due to human intervention. In case we cannot find the
			// node we assume the draining was successful and set the drainer config
			// status accordingly.

			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster node not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "setting drainer config status of tenant cluster node to drained condition")

			drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewDrainedCondition())

			_, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(drainerConfig.GetNamespace()).UpdateStatus(ctx, &drainerConfig, metav1.UpdateOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "set drainer config status of tenant cluster node to drained condition")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoned tenant cluster node")
	}

	var customPods []v1.Pod
	var systemPods []v1.Pod
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for all pods running on the tenant cluster node")

		fieldSelector := fields.SelectorFromSet(fields.Set{
			"spec.nodeName": key.NodeNameFromDrainerConfig(drainerConfig),
		})
		listOptions := metav1.ListOptions{
			FieldSelector: fieldSelector.String(),
		}
		podList, err := k8sClient.CoreV1().Pods(v1.NamespaceAll).List(ctx, listOptions)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, p := range podList.Items {
			if key.IsCriticalPod(p.Name) {
				// ignore critical pods (api, controller-manager and scheduler)
				// they are static pods so kubelet will recreate them anyway and it can cause other issues
				continue
			}
			if key.IsDaemonSetPod(p) {
				// ignore daemonSet owned pods
				// daemonSets pod are recreated even on unschedulable node so draining doesn't make sense
				// we are aligning here with community as 'kubectl drain' also ignore them
				continue
			}
			if key.IsEvictedPod(p) {
				// we don't need to care about already evicted pods
				continue
			}

			if p.GetNamespace() == "kube-system" {
				systemPods = append(systemPods, p)
			} else {
				customPods = append(customPods, p)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d pods running custom workloads", len(customPods)))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d pods running system workloads", len(systemPods)))
	}

	if len(customPods) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "sending eviction to all pods running custom workloads")

		for _, p := range customPods {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("sending eviction to pod %#q", fmt.Sprintf("%s/%s", p.GetNamespace(), p.GetName())))

			err := evictPod(ctx, k8sClient, p)
			if IsCannotEvictPod(err) {
				r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("cannot evict pod %#q due to disruption budget", p.GetName()))
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("sent eviction to pod %#q", fmt.Sprintf("%s/%s", p.GetNamespace(), p.GetName())))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "sent eviction to all pods running custom workloads")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pods running custom workloads to send evictions to")
	}

	// evict systemPods after all customPods are evicted
	if len(systemPods) > 0 && len(customPods) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "sending eviction to all pods running system workloads")

		for _, p := range systemPods {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("sending eviction to pod %#q", fmt.Sprintf("%s/%s", p.GetNamespace(), p.GetName())))

			err := evictPod(ctx, k8sClient, p)
			if IsCannotEvictPod(err) {
				r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("cannot evict pod %#q due to disruption budget", p.GetName()))
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("sent eviction to pod %#q", fmt.Sprintf("%s/%s", p.GetNamespace(), p.GetName())))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "sent eviction to all pods running system workloads")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pods running system workloads to send evictions to")
	}

	// When all pods are evicted from the tenant node, set the CR status to drained.
	if len(systemPods) == 0 && len(customPods) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting drainer config status of node in tenant cluster '%s' to drained condition", key.ClusterIDFromDrainerConfig(drainerConfig)))

		drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewDrainedCondition())

		_, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(drainerConfig.GetNamespace()).UpdateStatus(ctx, &drainerConfig, metav1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set drainer config status of node in tenant cluster '%s' to drained condition", key.ClusterIDFromDrainerConfig(drainerConfig)))
	}

	return nil
}

func drainingTimedOut(drainerConfig v1alpha1.DrainerConfig, now time.Time, timeout time.Duration) bool {
	return !drainerConfig.GetCreationTimestamp().Add(timeout).After(now)
}
