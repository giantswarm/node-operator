package healthz

import (
	"github.com/giantswarm/k8shealthz"
	"github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Service is the healthz service collection.
type Service struct {
	K8s healthz.Service
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	var err error

	var k8sService healthz.Service
	{
		c := k8shealthz.Config{}

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		k8sService, err = k8shealthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		K8s: k8sService,
	}

	return newService, nil
}
