package service

import (
	"github.com/giantswarm/operatorkit/v8/pkg/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
