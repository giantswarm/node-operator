package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v2/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for a service account in the Kubernetes API")

	namespace := key.ClusterNamespace(customObject)
	var currentServiceAccount *apiv1.ServiceAccount
	currentServiceAccount, err = r.k8sClient.CoreV1().ServiceAccounts(namespace).Get(key.ServiceAccountName(customObject), apismetav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "debug", "did not find the service account in the Kubernetes API")
		//when is not found api still returning non nil so it fails the create/update actions
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "debug", "found a service account in the Kubernetes API")
	}

	return currentServiceAccount, nil
}