package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/node-operator/service/controller/v1"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v1.VersionBundle())

	return versionBundles
}
