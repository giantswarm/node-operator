package service

import (
	"github.com/giantswarm/operatorkit/v7/pkg/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
