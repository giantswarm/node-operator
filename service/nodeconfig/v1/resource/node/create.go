package node

import (
	"context"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {

	// TODO fetch guest cluster certs
	// TODO create k8s client for guest cluster
	// TODO set guest cluster node unschedulable
	// TODO fetch all pods running on guest cluster node
	// TODO delete all pods running on guest cluster node
	// TODO delete CRO

	return nil
}
