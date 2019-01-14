package v2

import (
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
		Name:    "node-operator",
		Version: "0.1.0",
	}
}
