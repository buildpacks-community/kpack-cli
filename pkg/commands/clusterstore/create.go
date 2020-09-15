// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"io"

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	var (
		buildpackages []string
		dryRunConfig  DryRunConfig
	)

	cmd := &cobra.Command{
		Use:   "create <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Create a cluster store",
		Long: `Create a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

This clusterstore will be created only if it does not exist.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore create my-store -b my-registry.com/my-buildpackage
kp clusterstore create my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage
kp clusterstore create my-store -b ../path/to/my-local-buildpackage.cnb`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if dryRunConfig.dryRun {
				factory.Printer = commands.NewDiscardPrinter()
			} else {
				factory.Printer = commands.NewPrinter(cmd)
			}
			dryRunConfig.writer = cmd.OutOrStdout()

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			return create(name, buildpackages, factory, dryRunConfig, cs)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	cmd.Flags().BoolVarP(&dryRunConfig.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVarP(&dryRunConfig.outputFormat, "output", "o", "yaml", "output format. supported formats are: yaml, json")
	return cmd
}

type DryRunConfig struct {
	dryRun       bool
	outputFormat string
	writer       io.Writer
}

func create(name string, buildpackages []string, factory *clusterstore.Factory, drc DryRunConfig, cs k8s.ClientSet) (err error) {
	factory.Repository, err = k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	factory.Printer.Printf("Creating Cluster Store...")
	newStore, err := factory.MakeStore(name, buildpackages...)
	if err != nil {
		return err
	}

	if drc.dryRun {
		printer, err := commands.NewResourcePrinter(drc.outputFormat)
		if err != nil {
			return err
		}

		return printer.PrintObject(newStore, drc.writer)
	}

	_, err = cs.KpackClient.KpackV1alpha1().ClusterStores().Create(newStore)
	if err != nil {
		return err
	}

	factory.Printer.Printf("\"%s\" created", newStore.Name)
	return nil
}
