package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/node-operator/service/nodeconfig/v1/key"
)

const (
	// UnschedulablePatch is the JSON patch structure being applied to nodes using
	// a strategic merge patch in order to drain them.
	UnschedulablePatch = `{"spec":{"unschedulable":true}}`
)

// EnsureCreated represents the node resource implementation to manage on demand
// node draining for guest clusters.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var draining certs.Draining
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for certificates for the guest cluster")

		draining, err = r.certsSearcher.SearchDraining(key.ClusterID(customObject))
		if certs.IsTimeout(err) {
			// Here we log a warning for alerting purposes and also return an error to
			// make the resource execution being retried. Then the amount of warning
			// logs will surge and we have a chance to try to drain again in case
			// there are only some weird connection issues to the guest cluster
			// Kubernetes API.
			r.logger.LogCtx(ctx, "level", "warning", "message", "cannot find certificates for guest cluster '%s'", key.ClusterID(customObject))
			return microerror.Mask(err)
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found certificates for the guest cluster")
	}

	var k8sClient kubernetes.Interface
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes client for the guest cluster")

		var restConfig *rest.Config
		{
			c := k8srestconfig.Config{
				Logger: r.logger,

				Address:   key.ClusterAPIEndpoint(customObject),
				InCluster: false,
				TLS: k8srestconfig.TLSClientConfig{
					CAData:  draining.NodeOperator.CA,
					CrtData: draining.NodeOperator.Crt,
					KeyData: draining.NodeOperator.Key,
				},
			}

			restConfig, err = k8srestconfig.New(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		k8sClient, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created Kubernetes client for the guest cluster")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoning guest cluster node")

		n := key.NodeName(customObject)
		t := types.StrategicMergePatchType
		p := []byte(UnschedulablePatch)

		_, err := k8sClient.CoreV1().Nodes().Patch(n, t, p)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoned guest cluster node")
	}

	var customPods []v1.Pod
	var systemPods []v1.Pod
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for all pods running on the guest cluster node")

		fieldSelector := fields.SelectorFromSet(fields.Set{
			"spec.nodeName": key.NodeName(customObject),
		})
		listOptions := apismetav1.ListOptions{
			FieldSelector: fieldSelector.String(),
		}
		podList, err := k8sClient.CoreV1().Pods(v1.NamespaceAll).List(listOptions)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, p := range podList.Items {
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting all pods running custom workloads")

		for _, p := range customPods {
			err := k8sClient.CoreV1().Pods(p.GetNamespace()).Delete(p.GetName(), &apismetav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted all pods running custom workloads")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pods to be deleted running custom workloads")
	}

	if len(systemPods) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting all pods running system workloads")

		for _, p := range systemPods {
			err := k8sClient.CoreV1().Pods(p.GetNamespace()).Delete(p.GetName(), &apismetav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted all pods running system workloads")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pods to be deleted running system workloads")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "setting node config status of guest cluster node to final state")

		n := v1.NamespaceDefault
		c := v1alpha1.NodeConfigStatusCondition{
			Status: "True",
			Type:   "Drained",
		}
		customObject.Status.Conditions = append(customObject.Status.Conditions, c)

		_, err := r.g8sClient.CoreV1alpha1().NodeConfigs(n).Update(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "set node config status of guest cluster node to final state")
	}

	return nil
}
