package controller

import (
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/internal/selector"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/node-operator/pkg/project"
	"github.com/giantswarm/node-operator/service/controller/key"
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

	var resourceSet []resource.Interface
	{
		resourceSet, err = NewDrainerResourceSet(DrainerResourceSetConfig(config))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		resourceSets := [][]resource.Interface{
			resourceSet,
		}

		resources := []resource.Interface{}
		for _, set := range resourceSets {
			resources = append(resources, set...)
		}

		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			Resources:    resources,
			ResyncPeriod: 2 * time.Minute,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.DrainerConfig)
			},
			Selector: selector.NewSelector(
				key.LabelsDoNotIncludeNodeOperatorVersion,
			),

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
