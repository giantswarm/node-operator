package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/healthz"
	"github.com/giantswarm/microendpoint/endpoint/version"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/node-operator/server/middleware"
	"github.com/giantswarm/node-operator/service"
)

// Config represents the configuration used to create a endpoint.
type Config struct {
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// Endpoint is the endpoint collection.
type Endpoint struct {
	Healthz *healthz.Endpoint
	Version *version.Endpoint
}

// New creates a new configured endpoint.
func New(config Config) (*Endpoint, error) {
	var err error

	var healthzEndpoint *healthz.Endpoint
	{
		c := healthz.Config{}

		c.Logger = config.Logger
		c.Services = []healthzservice.Service{
			config.Service.Healthz.K8s,
		}

		healthzEndpoint, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionEndpoint *version.Endpoint
	{
		c := version.Config{}

		c.Logger = config.Logger
		c.Service = config.Service.Version

		versionEndpoint, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newEndpoint := &Endpoint{
		Healthz: healthzEndpoint,
		Version: versionEndpoint,
	}

	return newEndpoint, nil
}
