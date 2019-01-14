package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/node-operator/service/controller/v2/key"
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

	if customObject.Status.HasFinalCondition() {
		r.logger.LogCtx(ctx, "level", "debug", "message", "node config status already has final state")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource for custom object")

		return nil
	}

	k8sClient, err := r.guestCluster.NewK8sClient(ctx, key.ClusterID(customObject), key.ClusterAPIEndpoint(customObject))
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoning guest cluster node")

		n := key.NodeName(customObject)
		t := types.StrategicMergePatchType
		p := []byte(UnschedulablePatch)

		_, err := k8sClient.CoreV1().Nodes().Patch(n, t, p)
		if guest.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster API is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if apierrors.IsNotFound(err) {
			// It might happen the node we want to drain got already removed. This
			// might even be due to human intervention. In case we cannot find the
			// node we assume the draining was successful and set the node config
			// status accordingly.

			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster node not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "setting node config status of guest cluster node to final state")

			customObject.Status.Conditions = append(customObject.Status.Conditions, customObject.Status.NewFinalCondition())

			_, err := r.g8sClient.CoreV1alpha1().NodeConfigs(customObject.GetNamespace()).UpdateStatus(&customObject)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "set node config status of guest cluster node to final state")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource for custom object")

			return nil
		} else if err != nil {
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
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting node config status of node in guest cluster '%s' to final state", key.ClusterID(customObject)))

		customObject.Status.Conditions = append(customObject.Status.Conditions, customObject.Status.NewFinalCondition())

		_, err := r.g8sClient.CoreV1alpha1().NodeConfigs(customObject.GetNamespace()).UpdateStatus(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set node config status of node in guest cluster '%s' to final state", key.ClusterID(customObject)))
	}

	return nil
}
