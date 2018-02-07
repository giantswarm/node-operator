package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/microkit/transaction"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
	"github.com/giantswarm/node-operator/flag"
	"github.com/spf13/viper"

	"github.com/giantswarm/node-operator/server"
	"github.com/giantswarm/node-operator/service"
)

var (
	description = "Project node-operator drains Kubernetes nodes on behalf of watched CRDs."
	f           = flag.New()
	gitCommit   = "n/a"
	name        = "node-operator"
	source      = "https://github.com/giantswarm/node-operator"
)

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", microerror.Mask(err)))
	}
}

func mainWithError() error {
	var err error

	// Create a new logger which is used by all packages.
	var newLogger micrologger.Logger
	{
		c := micrologger.DefaultConfig()
		c.IOWriter = os.Stdout

		newLogger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{}

			c.Flag = f
			c.Logger = newLogger
			c.Viper = v

			c.Description = description
			c.GitCommit = gitCommit
			c.Name = name
			c.Source = source

			newService, err = service.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", err))
			}
			go newService.Boot()
		}

		var storage microstorage.Storage
		{
			storage, err = memory.New(memory.DefaultConfig())
			if err != nil {
				panic(fmt.Sprintf("%#v", err))
			}
		}

		var transactionResponder transaction.Responder
		{
			c := transaction.DefaultResponderConfig()
			c.Logger = newLogger
			c.Storage = storage

			transactionResponder, err = transaction.NewResponder(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", err))
			}
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.DefaultConfig()

			c.MicroServerConfig.Logger = newLogger
			c.MicroServerConfig.TransactionResponder = transactionResponder
			c.MicroServerConfig.ServiceName = name
			c.MicroServerConfig.Viper = v
			c.Service = newService

			newServer, err = server.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", err))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.DefaultConfig()

		c.Logger = newLogger
		c.ServerFactory = newServerFactory

		c.Description = description
		c.GitCommit = gitCommit
		c.Name = name
		c.Source = source
		c.VersionBundles = service.NewVersionBundles()

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	newCommand.CobraCommand().Execute()

	return nil
}
