package node

import (
	"context"
)

// EnsureDeleted is a noop, because the node resource implementation is not
// interested in delete events of the nodeconfigs.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
