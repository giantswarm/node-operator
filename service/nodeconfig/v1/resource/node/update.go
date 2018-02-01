package node

import (
	"context"

	"github.com/giantswarm/operatorkit/framework"
)

// ApplyUpdateChange is a noop. Draining nodes does not require more complex
// CRUD management of resources. See ApplyCreateChange for the business logic
// implementation.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	patch := framework.NewPatch()
	patch.SetCreateChange(true) // hack to execute ApplyCreateChange
	return patch, nil
}
