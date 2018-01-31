package node

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "nodev1"
)

type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(c Config) (*Resource, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	r := &Resource{
		k8sClient: c.K8sClient,
		logger: c.Logger.With(
			"resource", Name,
		),
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
