package v1

import (
	"context"
	"time"

	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/micrologger/loggermeta"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"

	"github.com/giantswarm/node-operator/service/controller/v1/key"
	"github.com/giantswarm/node-operator/service/controller/v1/resource/drainer"
)

type DrainerResourceSetConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,

			CertID: certs.NodeOperatorCert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerResource resource.Interface
	{
		c := drainer.Config{
			G8sClient:     config.K8sClient.G8sClient(),
			Logger:        config.Logger,
			TenantCluster: tenantCluster,
		}

		drainerResource, err = drainer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		drainerResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		drainerConfig, err := key.ToDrainerConfig(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersionFromDrainerConfig(drainerConfig) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		drainerConfig, err := key.ToDrainerConfig(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		lm := loggermeta.New()
		lm.KeyVals["cluster"] = key.ClusterIDFromDrainerConfig(drainerConfig)
		lm.KeyVals["node"] = key.NodeNameFromDrainerConfig(drainerConfig)

		return loggermeta.NewContext(ctx, lm), nil
	}

	var drainerResourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		drainerResourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return drainerResourceSet, nil
}
