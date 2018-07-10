package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
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
