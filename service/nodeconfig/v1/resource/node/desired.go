package node

import (
	"context"
)

// GetDesiredState is a noop. Draining nodes does not require more complex CRUD
// management of resources. See ApplyCreateChange for the business logic
// implementation.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}
