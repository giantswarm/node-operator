package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ToCustomObject(v interface{}) (v1alpha1.NodeConfig, error) {
	p, ok := v.(*v1alpha1.NodeConfig)
	if !ok {
		return v1alpha1.NodeConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.NodeConfig{}, v)
	}
	o := *p

	return o, nil
}

func VersionBundleVersion(customObject v1alpha1.NodeConfig) string {
	return customObject.Spec.VersionBundle.Version
}