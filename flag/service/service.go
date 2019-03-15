package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
