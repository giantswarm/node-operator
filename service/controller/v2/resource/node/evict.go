package node

import (
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	waitInterval = time.Millisecond * 500 // check every 500ms
	waitTimeout  = time.Minute * 2        // timeout after 2 min
)

// evict pod from node
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
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		DeleteOptions: deleteOptions,
	}
	// evict pod
	err := k8sClient.PolicyV1beta1().Evictions(eviction.Namespace).Evict(eviction)
	if err != nil {
		return microerror.Mask(err)
	}

	getOpts := apismetav1.GetOptions{}
	// wait for successful termination
	b := backoff.NewConstant(waitTimeout, waitInterval)
	o := func() error {
		p, err := k8sClient.CoreV1().Pods(pod.Namespace).Get(pod.Name, getOpts)
		if apierrors.IsNotFound(err) || (p != nil && p.ObjectMeta.UID != pod.ObjectMeta.UID) {
			// pod is no longer in api, we can exit
			return nil
		} else if err != nil {
			return err
		}
		return podNotTerminatedError
	}
	err = backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
