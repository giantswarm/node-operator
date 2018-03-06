package service

import (
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/node-operator/flag"
	"github.com/giantswarm/node-operator/service/healthz"
	"github.com/giantswarm/node-operator/service/nodeconfig"
)

type Config struct {
	Logger micrologger.Logger

	Description string
	Flag        *flag.Flag
	GitCommit   string
	Name        string
	Source      string
	Viper       *viper.Viper
}

type Service struct {
	Healthz             *healthz.Service
	NodeConfigFramework *framework.Framework
	Version             *version.Service

	bootOnce sync.Once
}

func New(config Config) (*Service, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var healthzService *healthz.Service
	{
		c := healthz.Config{}

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		healthzService, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodeConfigFramework *framework.Framework
	{
		c := nodeconfig.FrameworkConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			ProjectName: config.Name,
		}

		nodeConfigFramework, err = nodeconfig.NewFramework(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.DefaultConfig()

		c.Description = config.Description
		c.GitCommit = config.GitCommit
		c.Name = config.Name
		c.Source = config.Source
		c.VersionBundles = NewVersionBundles()

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Healthz:             healthzService,
		NodeConfigFramework: nodeConfigFramework,
		Version:             versionService,

		bootOnce: sync.Once{},
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.NodeConfigFramework.Boot()
	})
}
