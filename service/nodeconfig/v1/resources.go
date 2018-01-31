package v1

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/node-operator/service/nodeconfig/v1/resource/node"
)

const (
	ResourceRetries uint64 = 3
)

type ResourcesConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	Name string
}

func NewResources(config ResourcesConfig) ([]framework.Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Name must not be empty", config)
	}

	var err error

	var nodeResource framework.Resource
	{
		c := node.Config{}

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
		c := retryresource.DefaultWrapConfig()

		c.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		c.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.DefaultWrapConfig()

		c.Name = config.Name

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
