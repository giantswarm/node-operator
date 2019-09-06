package service

import (
	"context"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/node-operator/flag"
	"github.com/giantswarm/node-operator/service/controller"
)

type Config struct {
	Flag   *flag.Flag
	Logger micrologger.Logger
	Viper  *viper.Viper

	Description string
	GitCommit   string
	Name        string
	Source      string
	Version     string
}

type Service struct {
	Version *version.Service

	bootOnce          sync.Once
	drainerController *controller.Drainer
}

func New(config Config) (*Service, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag must not be empty", config)
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Viper must not be empty", config)
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:    config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

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

	var drainerController *controller.Drainer
	{
		c := controller.DrainerConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			ProjectName: config.Name,
		}

		drainerController, err = controller.NewDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description:    config.Description,
			GitCommit:      config.GitCommit,
			Name:           config.Name,
			Source:         config.Source,
			Version:        config.Version,
			VersionBundles: NewVersionBundles(),
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Version: versionService,

		bootOnce:          sync.Once{},
		drainerController: drainerController,
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.drainerController.Boot(context.Background())
	})
}
