package node

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "nodev1"
)

type Config struct {
	CertsSearcher certs.Interface
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
}

type Resource struct {
	certsSearcher certs.Interface
	g8sClient     versioned.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
}

func New(c Config) (*Resource, error) {
	if c.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", c)
	}
	if c.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", c)
	}
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	r := &Resource{
		certsSearcher: c.CertsSearcher,
		g8sClient:     c.G8sClient,
		k8sClient:     c.K8sClient,
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
