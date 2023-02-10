package drainer

import (
	"context"
	"fmt"
	"os"
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	v1alpha1 "github.com/giantswarm/node-operator/api"
	"github.com/giantswarm/tenantcluster/v5/pkg/tenantcluster"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
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

	// Get AWSCluster Object to write events on it
	awsCluster := &infrastructurev1alpha3.AWSCluster{}
	err = r.client.Get(ctx, types.NamespacedName{Name: key.ClusterIDFromDrainerConfig(drainerConfig), Namespace: drainerConfig.Namespace}, awsCluster)
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
			Timeout:                         2 * time.Minute, // give a 5 minutes timeout
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

			typeOfNode := "worker"
			// In case of master nodes, just delete the pods, and don't evict them
			if nodeIsMaster(&node) {

				// For the master node wait 30 seconds for a pod to be terminated
				nodeShutdownHelper.GracePeriodSeconds = 30

				// We want the master nodes to be terminated rather quickly
				nodeShutdownHelper.Timeout = 1 * time.Minute

				// Set type to master
				typeOfNode = "master"
			}

			// Check if we are done with the draining of the specific node
			if draining, ok := r.draining[nodeName]; ok {
				select {
				case drainingError := <-draining:

					// Set the timeout condition
					drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewTimeoutCondition())

					// if it succeeded then remove the entry for the specific node from the state
					if err := r.client.Status().Update(ctx, &drainerConfig); err == nil {
						delete(r.draining, nodeName)
					} else {
						// otherwise re-queue the error so that we can retry again
						draining <- drainingError
					}

				case <-time.After(5 * time.Second):
					// we want to wait only for a max of N seconds, otherwise continue
				}

			} else {

				// drain async and add the status to the state
				// Important run in a different go routine
				go r.drainNode(nodeName, typeOfNode, ctx, *awsCluster, nodeShutdownHelper, node, k8sClient, drainerConfig)

			}

			break
		}
	}

	return nil
}

func (r *Resource) drainNode(
	nodeName string,
	typeOfNode string,
	ctx context.Context,
	awsCluster infrastructurev1alpha3.AWSCluster,
	shutdownHelper drain.Helper,
	node v1.Node, k8sClient kubernetes.Interface,
	drainerConfig v1alpha1.DrainerConfig) {

	// Await channel
	await := make(chan error, 1)

	// Cordon the node
	if err := drain.RunCordonOrUncordon(&shutdownHelper, &node, true); err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to cordon %s node with error %s", typeOfNode, err))
		r.event.Warn(ctx, &awsCluster, "CordoningFailed", fmt.Sprintf("failed to cordon %s node %s with error %s", typeOfNode, node.GetName(), err))
		return
	}

	// Log the node as cordoned
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cordoned %s node", typeOfNode))

	r.lock.Lock()
	r.draining[nodeName] = await
	r.lock.Unlock()

	// It means the cordoning was successful, proceed with the draining
	// The draining function is going to block until the draining is successful
	// or a timeout happens (whichever happens first)
	if err := drain.RunNodeDrain(&shutdownHelper, nodeName); err != nil {

		// This means the draining failed
		// Log it
		r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to drain %s node with error %s", typeOfNode, err))
		r.event.Warn(ctx, &awsCluster, "DrainingFailed", fmt.Sprintf("failed to drain %s node %s with error %s", typeOfNode, node.GetName(), err))

		// log all the pods that could not be evicted or deleted
		r.logUnevictedPods(k8sClient, ctx, &awsCluster, typeOfNode, &node)

		// now set the timeout condition, which means the aws-operator will proceed to delete the node
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting drainer config status of tenant cluster %s node to timeout condition", typeOfNode))

		// Set the timeout condition
		drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewTimeoutCondition())

		// Apply it to the resource in k8s
		err := r.client.Status().Update(ctx, &drainerConfig)

		if err != nil {
			await <- microerror.Mask(err)
			return
		}

	} else {

		// if we got here it means we have got no errors and the node is successfully drained
		// Set the drainer status in the node so that the aws-operator can proceed with the deletion
		// of the node
		drainerConfig.Status.Conditions = append(drainerConfig.Status.Conditions, drainerConfig.Status.NewDrainedCondition())

		// Now update the node status
		err := r.client.Status().Update(ctx, &drainerConfig)
		if err != nil {
			await <- microerror.Mask(err)
			return
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set drainer config status of tenant cluster %s node to drained condition", typeOfNode))
		r.event.Info(ctx, &awsCluster, "DrainingSucceeded", fmt.Sprintf("drained %s node %s successfully", typeOfNode, node.GetName()))
	}
	await <- nil
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

func (r *Resource) logUnevictedPods(k8sClient kubernetes.Interface, ctx context.Context, awsCluster *infrastructurev1alpha3.AWSCluster, typeOfNode string, node *v1.Node) {
	// Get the list of pods for the specific node
	nodePods, err := nodePods(k8sClient, ctx, node)

	// Log all the pods that could not be drained/deleted
	if err == nil {
		for _, pod := range nodePods {
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("could not evict/delete pod %s on %s node", pod.GetName(), typeOfNode))
			r.event.Warn(ctx, awsCluster, "DrainerConfigFailed", fmt.Sprintf("%s node %s could not evict/delete pod %s", typeOfNode, node.GetName(), pod.GetName()))
		}
	}

	// if instead we got an error log it
	if err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("could not get the list of pods: %s", err))
		r.event.Warn(ctx, awsCluster, "DrainerConfigFailed", fmt.Sprintf("could not get the list of pods for the node %s: %s", node.GetName(), err))
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
