package v1

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/node-operator/service/nodeconfig/v1/key"
	"github.com/giantswarm/node-operator/service/nodeconfig/v1/resource/node"
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

func NewResourceSet(config ResourceSetConfig) (*framework.ResourceSet, error) {
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var err error

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodeResource framework.Resource
	{
		c := node.Config{}

		c.CertsSearcher = certsSearcher
		c.G8sClient = config.G8sClient
		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		nodeResource, err = node.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		nodeResource,
	}

	{
		c := retryresource.WrapConfig{}

		c.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		c.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		c.Name = config.ProjectName

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

	var resourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{}

		c.Handles = handlesFunc
		c.Logger = config.Logger
		c.Resources = resources

		resourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
