package node

import (
	"context"

	"github.com/giantswarm/operatorkit/framework"
)

// ApplyDeleteChange is a noop. Draining nodes does not require more complex
// CRUD management of resources. See ApplyCreateChange for the business logic
// implementation.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// NewDeletePatch is a noop. Draining nodes does not require more complex CRUD
// management of resources. See ApplyCreateChange for the business logic
// implementation.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	patch := framework.NewPatch()
	return patch, nil
}
