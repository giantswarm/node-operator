package controller

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/node-operator/pkg/project"
	v1 "github.com/giantswarm/node-operator/service/controller/v1"
	v2 "github.com/giantswarm/node-operator/service/controller/v2"
)

type DrainerConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Drainer struct {
	*controller.Controller
}

func NewDrainer(config DrainerConfig) (*Drainer, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	var v1ResourceSet *controller.ResourceSet
	{
		c := v1.DrainerResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		v1ResourceSet, err = v1.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v2ResourceSet *controller.ResourceSet
	{
		c := v2.DrainerResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		v2ResourceSet, err = v2.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
				v2ResourceSet,
			},
			ResyncPeriod: 2 * time.Minute,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.DrainerConfig)
			},

			Name: project.Name(),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &Drainer{
		Controller: operatorkitController,
	}

	return d, nil
}
