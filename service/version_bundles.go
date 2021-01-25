package service

import (
	"github.com/giantswarm/node-operator/service/controller"
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, controller.VersionBundle())

	return versionBundles
}
