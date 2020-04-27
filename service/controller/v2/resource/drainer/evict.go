package drainer

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func evictPod(k8sClient kubernetes.Interface, pod v1.Pod) error {
	eviction := &v1beta1.Eviction{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
		DeleteOptions: &apismetav1.DeleteOptions{
			GracePeriodSeconds: terminationGracePeriod(pod),
		},
	}

	err := k8sClient.PolicyV1beta1().Evictions(eviction.GetNamespace()).Evict(eviction)
	if IsCannotEvictPod(err) {
		return microerror.Mask(cannotEvictPodError)
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func terminationGracePeriod(pod v1.Pod) *int64 {
	var d int64 = 60

	if pod.Spec.TerminationGracePeriodSeconds != nil && *pod.Spec.TerminationGracePeriodSeconds > 0 {
		d = *pod.Spec.TerminationGracePeriodSeconds
	}

	return to.Int64P(d)
}
