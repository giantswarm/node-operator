package drainer

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var cannotEvictPodError = &microerror.Error{
	Kind: "cannotEvictPodError",
}

// IsCannotEvictPod asserts cannotEvictPodError.
func IsCannotEvictPod(err error) bool {
	c := microerror.Cause(err)

	if err == nil {
		return false
	}

	if strings.Contains(c.Error(), "Cannot evict pod") {
		return true
	}

	return c == cannotEvictPodError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
