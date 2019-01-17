package drainer

import (
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func EvictPod(k8sClient kubernetes.Interface, pod v1.Pod) error {
	var deleteGracePeriod int64 = 60
	if pod.DeletionGracePeriodSeconds != nil && *pod.DeletionGracePeriodSeconds > 0 {
		deleteGracePeriod = *pod.DeletionGracePeriodSeconds
	}
	deleteOptions := &apismetav1.DeleteOptions{
		GracePeriodSeconds: &deleteGracePeriod,
	}
	eviction := &v1beta1.Eviction{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
		DeleteOptions: deleteOptions,
	}
	err := k8sClient.PolicyV1beta1().Evictions(eviction.GetNamespace()).Evict(eviction)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
