package drainer

import (
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v6/pkg/tenantcluster"
	"sigs.k8s.io/controller-runtime/pkg/client"

	event "github.com/giantswarm/node-operator/service/recorder"
)

const (
	Name = "drainerv2"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

type Config struct {
	Client        client.Client
	Event         event.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

type NodeName = string

type Resource struct {
	client        client.Client
	event         event.Interface
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface

	lock     sync.RWMutex
	draining map[NodeName]chan error
}

func New(c Config) (*Resource, error) {
	if c.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", c)
	}

	r := &Resource{
		client:        c.Client,
		event:         c.Event,
		logger:        c.Logger,
		tenantCluster: c.TenantCluster,
		lock:          sync.RWMutex{},
		draining:      make(map[string]chan error),
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
