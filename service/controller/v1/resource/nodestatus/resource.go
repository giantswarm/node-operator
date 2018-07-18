package nodestatus

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/guestcluster"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "nodestatusv1"
)

type Config struct {
	GuestCluster guestcluster.Interface
	G8sClient    versioned.Interface
	Logger       micrologger.Logger
}

type Resource struct {
	guestCluster guestcluster.Interface
	g8sClient    versioned.Interface
	logger       micrologger.Logger
}

func New(c Config) (*Resource, error) {
	if c.GuestCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestCluster must not be empty", c)
	}
	if c.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	r := &Resource{
		guestCluster: c.GuestCluster,
		g8sClient:    c.G8sClient,
		logger:       c.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
