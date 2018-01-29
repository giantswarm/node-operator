package service

import (
	"github.com/giantswarm/node-operator/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
