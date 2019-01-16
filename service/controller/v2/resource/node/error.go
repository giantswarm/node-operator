package node

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var podNotTerminatedError = &microerror.Error{
	Kind: "podNotTerminatedError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsPodNotTerminatedError(err error) bool {
	return microerror.Cause(err) == podNotTerminatedError
}
