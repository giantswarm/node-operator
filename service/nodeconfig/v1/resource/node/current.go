package node

import (
	"context"
)

// GetCurrentState is a noop. Draining nodes does not require more complex CRUD
// management of resources. See ApplyCreateChange for the business logic
// implementation.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}
