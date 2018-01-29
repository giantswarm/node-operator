package middleware

import (
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/node-operator/service"
)

// Config represents the configuration used to create a middleware.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Service *service.Service
}

// Middleware is middleware collection.
type Middleware struct {
}

// New creates a new configured middleware.
func New(config Config) (*Middleware, error) {
	return &Middleware{}, nil
}
