package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/node-operator/service/controller/v1/key"
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

	clusterID := key.ClusterID(customObject)

	if customObject.Status.HasFinalCondition() {
		r.logger.LogCtx(ctx, "level", "debug", "message", "node config status already has final state", "clusterID", clusterID)
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource for custom object", "clusterID", clusterID)

		return nil
	}

	k8sClient, err := r.guestCluster.NewK8sClient(ctx, key.ClusterID(customObject), key.ClusterAPIEndpoint(customObject))
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoning guest cluster node", "clusterID", clusterID)

		n := key.NodeName(customObject)
		t := types.StrategicMergePatchType
		p := []byte(UnschedulablePatch)

		_, err := k8sClient.CoreV1().Nodes().Patch(n, t, p)
		if apierrors.IsNotFound(err) {
			// It might happen the node we want to drain got already removed. This
			// might even be due to human intervention. In case we cannot find the
			// node we assume the draining was successful and set the node config
			// status accordingly.

			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster node not found", "clusterID", clusterID)
			r.logger.LogCtx(ctx, "level", "debug", "message", "setting node config status of guest cluster node to final state", "clusterID", clusterID)

			customObject.Status.Conditions = append(customObject.Status.Conditions, customObject.Status.NewFinalCondition())

			_, err := r.g8sClient.CoreV1alpha1().NodeConfigs(customObject.GetNamespace()).Update(&customObject)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "set node config status of guest cluster node to final state", "clusterID", clusterID)
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource for custom object", "clusterID", clusterID)

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoned guest cluster node", "clusterID", clusterID)
	}

	var customPods []v1.Pod
	var systemPods []v1.Pod
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for all pods running on the guest cluster node", "clusterID", clusterID)

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

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d pods running custom workloads", len(customPods)), "clusterID", clusterID)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d pods running system workloads", len(systemPods)), "clusterID", clusterID)
	}

	if len(customPods) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting all pods running custom workloads", "clusterID", clusterID)

		for _, p := range customPods {
			err := k8sClient.CoreV1().Pods(p.GetNamespace()).Delete(p.GetName(), &apismetav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted all pods running custom workloads", "clusterID", clusterID)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pods to be deleted running custom workloads", "clusterID", clusterID)
	}

	if len(systemPods) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting all pods running system workloads", "clusterID", clusterID)

		for _, p := range systemPods {
			err := k8sClient.CoreV1().Pods(p.GetNamespace()).Delete(p.GetName(), &apismetav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted all pods running system workloads", "clusterID", clusterID)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pods to be deleted running system workloads", "clusterID", clusterID)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting node config status of node in guest cluster '%s' to final state", key.ClusterID(customObject)), "clusterID", clusterID)

		customObject.Status.Conditions = append(customObject.Status.Conditions, customObject.Status.NewFinalCondition())

		_, err := r.g8sClient.CoreV1alpha1().NodeConfigs(customObject.GetNamespace()).Update(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set node config status of node in guest cluster '%s' to final state", key.ClusterID(customObject)), "clusterID", clusterID)
	}

	return nil
}
