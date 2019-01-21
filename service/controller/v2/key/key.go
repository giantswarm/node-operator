package key

import (
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ClusterAPIEndpoint(customObject v1alpha1.NodeConfig) string {
	return customObject.Spec.Guest.Cluster.API.Endpoint
}

func ClusterEndpointFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.Guest.Cluster.API.Endpoint
}

func ClusterID(customObject v1alpha1.NodeConfig) string {
	return customObject.Spec.Guest.Cluster.ID
}

func ClusterIDFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.Guest.Cluster.ID
}

func IsCriticalPod(podName string) bool {
	r := false
	r = r || strings.HasPrefix(podName, "k8s-api-server")
	r = r || strings.HasPrefix(podName, "k8s-controller-manager")
	r = r || strings.HasPrefix(podName, "k8s-scheduler")

	return r
}

func IsDaemonSetPod(pod v1.Pod) bool {
	r := false
	ownerRefrence := metav1.GetControllerOf(&pod)

	if ownerRefrence != nil && ownerRefrence.Kind == "DaemonSet" {
		r = true
	}

	return r
}

func IsEvicted(pod v1.Pod) bool {
	return pod.Status.Reason == "Evicted"
}

func NodeName(customObject v1alpha1.NodeConfig) string {
	return customObject.Spec.Guest.Node.Name
}

func NodeNameFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.Guest.Node.Name
}

func ToCustomObject(v interface{}) (v1alpha1.NodeConfig, error) {
	p, ok := v.(*v1alpha1.NodeConfig)
	if !ok {
		return v1alpha1.NodeConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.NodeConfig{}, v)
	}
	o := *p

	return o, nil
}

func ToDrainerConfig(v interface{}) (v1alpha1.DrainerConfig, error) {
	p, ok := v.(*v1alpha1.DrainerConfig)
	if !ok {
		return v1alpha1.DrainerConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.DrainerConfig{}, v)
	}
	o := *p

	return o, nil
}

func VersionBundleVersion(customObject v1alpha1.NodeConfig) string {
	return customObject.Spec.VersionBundle.Version
}

func VersionBundleVersionFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.VersionBundle.Version
}
