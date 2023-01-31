package recorder

import (
	"context"

	pkgruntime "k8s.io/apimachinery/pkg/runtime"
)

type Interface interface {
	Info(ctx context.Context, obj pkgruntime.Object, reason, message string)
	Warn(ctx context.Context, obj pkgruntime.Object, reason, message string)
}
