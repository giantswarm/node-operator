package v1

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "node-operator",
				Description: "Introduce the first version of the node-operator.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "node-operator",
				Version: "0.1.0",
			},
		},
		Dependencies: []versionbundle.Dependency{
			{
				Name:    "kubernetes",
				Version: ">= 1.8.4",
			},
		},
		Deprecated: false,
		Name:       "node-operator",
		Time:       time.Date(2018, time.February, 6, 18, 6, 0, 0, time.UTC),
		Version:    "0.1.0",
		WIP:        true,
	}
}
