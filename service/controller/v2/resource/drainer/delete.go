package drainer

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/node-operator/service/controller/v2/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	drainerConfig, err := key.ToDrainerConfig(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var k8sClient kubernetes.Interface
	{
		i := key.ClusterIDFromDrainerConfig(drainerConfig)
		e := key.ClusterEndpointFromDrainerConfig(drainerConfig)
		restConfig, err := r.tenantCluster.NewRestConfig(ctx, i, e)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "fetching certificates timed out")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		clientsConfig := k8sclient.ClientsConfig{
			Logger:     r.logger,
			RestConfig: restConfig,
		}
		k8sClients, err := k8sclient.NewClients(clientsConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		k8sClient = k8sClients.K8sClient()
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting tenant cluster node from Kubernetes API")

		n := key.NodeNameFromDrainerConfig(drainerConfig)
		o := &metav1.DeleteOptions{}
		err := k8sClient.CoreV1().Nodes().Delete(n, o)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete tenant cluster node from Kubernetes API")
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster API is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete tenant cluster node from Kubernetes API")
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster node not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted tenant cluster node from Kubernetes API")
	}

	return nil
}
