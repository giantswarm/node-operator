package v2

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/guestcluster"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/micrologger/loggermeta"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/node-operator/service/controller/v2/key"
	"github.com/giantswarm/node-operator/service/controller/v2/resource/node"
)

const (
	ResourceRetries uint64 = 3
)

type ResourceSetConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	HandledVersionBundles []string
	ProjectName           string
}

func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var guestCluster guestcluster.Interface
	{
		c := guestcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,

			CertID: certs.NodeOperatorCert,
		}

		guestCluster, err = guestcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodeResource controller.Resource
	{
		c := node.Config{
			GuestCluster: guestCluster,
			G8sClient:    config.G8sClient,
			Logger:       config.Logger,
		}

		nodeResource, err = node.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		nodeResource,
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
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(customObject) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		nodeConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		lm := loggermeta.New()
		lm.KeyVals["cluster"] = key.ClusterID(nodeConfig)
		lm.KeyVals["node"] = key.NodeName(nodeConfig)

		return loggermeta.NewContext(ctx, lm), nil
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
