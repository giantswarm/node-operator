package drainer

import (
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
)

const (
	Name = "drainerv1"
)

type Config struct {
	G8sClient     versioned.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

type Resource struct {
	g8sClient     versioned.Interface
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface
}

func New(c Config) (*Resource, error) {
	if c.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", c)
	}

	r := &Resource{
		g8sClient:     c.G8sClient,
		logger:        c.Logger,
		tenantCluster: c.TenantCluster,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
