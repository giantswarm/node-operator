module github.com/giantswarm/node-operator

go 1.13

require (
	github.com/giantswarm/apiextensions/v3 v3.14.2-0.20210121112715-b6b39854aaef
	github.com/giantswarm/certs/v3 v3.1.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v4 v4.2.0
	github.com/giantswarm/tenantcluster/v4 v4.0.0
	github.com/giantswarm/to v0.3.0
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.18.15
	k8s.io/apimachinery v0.18.15
	k8s.io/client-go v0.18.15
)

replace sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
