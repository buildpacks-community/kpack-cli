// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

var (
	buildpackages []string
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <store> <buildpackage> [<buildpackage>...]",
		Short: "Create a cluster store",
		Long: `Create a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical registry.
Therefore, you must have credentials to access the registry on your machine.

This store will be created only if it does not exist.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp store create my-store my-registry.com/my-buildpackage
kp clusterstore create my-store --buildpackage my-registry.com/my-buildpackage --buildpackage my-registry.com/my-other-buildpackage
kp clusterstore create my-store --buildpackage ../path/to/my-local-buildpackage.cnb`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			factory.Printer = commands.NewPrinter(cmd)

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			factory.Repository, err = k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
			if err != nil {
				return err
			}

			newStore, err := factory.MakeStore(name, buildpackages...)
			if err != nil {
				return err
			}

			_, err = cs.KpackClient.ExperimentalV1alpha1().ClusterStores().Create(newStore)
			if err != nil {
				return err
			}

			factory.Printer.Printf("\"%s\" created", newStore.Name)
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	return cmd
}
