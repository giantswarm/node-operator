package service

import (
	"github.com/giantswarm/operatorkit/v4/pkg/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
