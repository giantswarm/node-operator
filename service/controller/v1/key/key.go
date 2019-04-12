package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ClusterEndpointFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.Guest.Cluster.API.Endpoint
}

func ClusterIDFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.Guest.Cluster.ID
}

func NodeNameFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.Guest.Node.Name
}

func ToDrainerConfig(v interface{}) (v1alpha1.DrainerConfig, error) {
	p, ok := v.(*v1alpha1.DrainerConfig)
	if !ok {
		return v1alpha1.DrainerConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.DrainerConfig{}, v)
	}
	o := *p

	return o, nil
}

func VersionBundleVersionFromDrainerConfig(drainerConfig v1alpha1.DrainerConfig) string {
	return drainerConfig.Spec.VersionBundle.Version
}
