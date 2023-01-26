package drainer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v5/pkg/tenantcluster"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/drain"

	"github.com/giantswarm/node-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	drainerConfig, err := key.ToDrainerConfig(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get the node name we want to cordon and drain
	nodeName := key.NodeNameFromDrainerConfig(drainerConfig)

	if drainerConfig.Status.HasDrainedCondition() {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%s drainer config status has drained condition", nodeName))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	if drainerConfig.Status.HasTimeoutCondition() {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%s drainer config status has timeout condition", nodeName))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	// ====================================================================
	// Setup the k8sclient

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

	// ====================================================================
	// Cordon and drain the node
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "cordoning tenant cluster node")

		// get the list of nodes
		nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

		// Check in case the k8s API is not available
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster API is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		// if we got here it means we have the list of nodes

		// WARNING
		// we are configuring here the draining behaviour for the worker nodes by default
		// however we will modify it for the master node in the node loop right below
		// WARNING
		nodeShutdownHelper := drain.Helper{
			Ctx:                             ctx,             // pass the current context
			Client:                          k8sClient,       // the k8s client for making the API calls
			Force:                           true,            // forcing the draining
			GracePeriodSeconds:              60,              // 60 seconds of timeout before deleting the pod
			IgnoreAllDaemonSets:             true,            // ignore the daemonsets
			Timeout:                         5 * time.Minute, // give a 5 minutes timeout
			DeleteEmptyDirData:              true,            // delete all the emptyDir volumes
			DisableEviction:                 false,           // we want to evict and not delete. (might be different for the master nodes)
			SkipWaitForDeleteTimeoutSeconds: 15,              // in case a node is NotReady then the pods won't be deleted, so don't wait too long
			Out:                             os.Stdout,
			ErrOut:                          os.Stderr,
			OnPodDeletedOrEvicted: func(pod *v1.Pod, usingEviction bool) {
				if pod != nil {
					if usingEviction {
						r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("evicted pod %s", pod.GetName()))
					} else {
						r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("deleted pod %s", pod.GetName()))
					}
				}
			},
		}

		// Loop through the list of nodes
		for _, node := range nodes.Items {

			node := node

			// If the node name does not match, simply continue
			if node.Name != nodeName {
				continue
			}

			// In case of master nodes, just delete the pods, and don't evict them
			if nodeIsMaster(&node) {

				// For the master node wait 30 seconds for a pod to be terminated
				nodeShutdownHelper.GracePeriodSeconds = 30

				// We want the master nodes to be terminated rather quickly
				nodeShutdownHelper.Timeout = 1 * time.Minute
			}

			// Cordon the node
			if err := drain.RunCordonOrUncordon(&nodeShutdownHelper, &node, true); err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to cordon node %s with error %s", node.GetName(), err))
			} else {

				// Log the node as cordoned
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cordoned tenant cluster node: %s", node.GetName()))

				// It means the cordoning was successful, proceed with the draining
				// The draining function is going to block until the draining is successful
				// or a timeout happens (whichever happens first)
				if err := drain.RunNodeDrain(&nodeShutdownHelper, nodeName); err != nil {

					// This means the draining failed
					// Log it
					r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to drain node %s with error %s", node.GetName(), err))

					// log all the pods that could not be evicted or deleted
					r.logUnevictedPods(k8sClient, ctx, &node)

					// now set the timeout condition, which means the aws-operator will proceed to delete the node
					r.logger.LogCtx(ctx, "level", "debug", "message",
						fmt.Sprintf("%s setting drainer config status of tenant cluster node to timeout condition", node.GetName()))

					drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewTimeoutCondition())

					err := r.client.Status().Update(ctx, &drainerConfig)
					if err != nil {
						return microerror.Mask(err)
					}

				} else {

					// if we got here it means we have got no errors and the node is successfully drained
					// Set the drainer status in the node so that the aws-operator can proceed with the deletion
					// of the node
					drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewDrainedCondition())

					// Now update the node status
					err := r.client.Status().Update(ctx, &drainerConfig)
					if err != nil {
						return microerror.Mask(err)
					}
					r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set drainer config status of tenant cluster node %s to drained condition", node.GetName()))

				}
			}
			break
		}
	}

	return nil
}

// Checks whether a node is a master node
func nodeIsMaster(node *v1.Node) bool {

	for key := range node.Labels {

		// New label
		if key == "node-role.kubernetes.io/control-plane" {
			return true
		}

		// Deprecated label
		if key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false

}

func (r *Resource) logUnevictedPods(k8sClient kubernetes.Interface, ctx context.Context, node *v1.Node) {
	// Get the list of pods for the specific node
	nodePods, err := nodePods(k8sClient, ctx, node)

	// Log all the pods that could not be drained/deleted
	if err == nil {
		for _, pod := range nodePods {
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("node %s could not evict/delete pod %s", node.GetName(), pod.GetName()))
		}
	}

	// if instead we got an error log it
	if err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("could not get the list of pods for the node %s: %s", node.GetName(), err))
	}
}

// Returns the list of pods for the node
func nodePods(k8sClient kubernetes.Interface, ctx context.Context, node *v1.Node) ([]v1.Pod, error) {
	fieldSelector := fields.SelectorFromSet(fields.Set{
		"spec.nodeName": node.GetName(),
	})
	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	}
	podList, err := k8sClient.CoreV1().Pods(v1.NamespaceAll).List(ctx, listOptions)

	return podList.Items, err
}
