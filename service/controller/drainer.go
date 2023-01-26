package controller

import (
	"fmt"
	"time"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/giantswarm/node-operator/api"
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

		// This selector selects DrainerConfigs where the node-operator version label is not
		// present in the given set of labels. This was added to allow node-operator to reconcile "old"
		// DrainerConfigs, which were versioned using their VersionBundle version, and prevent it from
		// reconciling possible future DrainerConfigs, which would be versioned using the label.
		// For more info, see https://github.com/giantswarm/giantswarm/issues/15423.
		selector, err := labels.Parse(fmt.Sprintf("!%s", key.LabelNodeOperatorVersion))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			Resources:    resources,
			ResyncPeriod: 1 * time.Minute,
			NewRuntimeObjectFunc: func() client.Object {
				return new(v1alpha1.DrainerConfig)
			},
			Selector: selector,

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
