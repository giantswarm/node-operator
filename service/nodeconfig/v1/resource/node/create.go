package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/azure-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// TODO fetch guest cluster certs
	draining, err := r.certsSearcher.SearchDraining(key.ClusterID(customObject))
	if err != nil {
		return microerror.Mask(err)
	}
	fmt.Printf("%#v\n", draining)

	// TODO create k8s client for guest cluster
	// TODO set guest cluster node unschedulable
	// TODO fetch all pods running on guest cluster node
	// TODO delete all pods running on guest cluster node
	// TODO delete CRO

	return nil
}
