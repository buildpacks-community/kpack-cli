// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	var (
		buildpackages []string
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
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			factory.Printer = ch

			return create(name, buildpackages, factory, ch, cs)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	commands.SetDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	return cmd
}

func create(name string, buildpackages []string, factory *clusterstore.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (err error) {
	factory.Repository, err = k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	if err = ch.PrintStatus("Creating ClusterStore..."); err != nil {
		return err
	}

	newStore, err := factory.MakeStore(name, buildpackages...)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		newStore, err = cs.KpackClient.KpackV1alpha1().ClusterStores().Create(newStore)
		if err != nil {
			return err
		}
	}

	if err = ch.PrintObj(newStore); err != nil {
		return err
	}

	return ch.PrintResult("ClusterStore %q created", newStore.Name)
}
