package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/node-operator/service/nodeconfig/v1/key"
)

const (
	UnschedulablePatch = `{"spec":{"unschedulable":true}}`
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	draining, err := r.certsSearcher.SearchDraining(key.ClusterID(customObject))
	if err != nil {
		return microerror.Mask(err)
	}

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = r.logger

		c.Address = key.ClusterAPIEndpoint(customObject)
		c.InCluster = false
		c.TLS.CAData = draining.NodeOperator.CA
		c.TLS.CrtData = draining.NodeOperator.Crt
		c.TLS.KeyData = draining.NodeOperator.Key

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		n := key.NodeName(customObject)
		t := types.StrategicMergePatchType
		p := []byte(UnschedulablePatch)

		{
			manifest, err := k8sClient.CoreV1().Nodes().Get(n, apismetav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			fmt.Printf("unschedulable\n")
			fmt.Printf("%#v\n", manifest.Spec.Unschedulable)
			fmt.Printf("unschedulable\n")
		}

		_, err := k8sClient.CoreV1().Nodes().Patch(n, t, p)
		if err != nil {
			return microerror.Mask(err)
		}

		{
			manifest, err := k8sClient.CoreV1().Nodes().Get(n, apismetav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			fmt.Printf("unschedulable\n")
			fmt.Printf("%#v\n", manifest.Spec.Unschedulable)
			fmt.Printf("unschedulable\n")
		}
	}

	// TODO set guest cluster node unschedulable
	// TODO fetch all pods running on guest cluster node
	// TODO delete all pods running on guest cluster node
	// TODO delete CRO

	return nil
}
