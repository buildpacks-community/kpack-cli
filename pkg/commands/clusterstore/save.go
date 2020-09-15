// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	var (
		buildpackages []string
		dryRunConfig DryRunConfig
	)

	cmd := &cobra.Command{
		Use:   "save <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Create or update a cluster store",
		Long: `Create or update a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

This clusterstore will be created only if it does not exist, otherwise it will be updated.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore save my-store -b my-registry.com/my-buildpackage
kp clusterstore save my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage
kp clusterstore save my-store -b ../path/to/my-local-buildpackage.cnb`,
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

			s, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(name, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return create(name, buildpackages, factory, dryRunConfig, cs)
			} else if err != nil {
				return err
			}

			return update(s, buildpackages, factory, dryRunConfig, cs)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	cmd.Flags().BoolVarP(&dryRunConfig.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVarP(&dryRunConfig.outputFormat, "output", "o", "yaml", "output format. supported formats are: yaml, json")
	return cmd
}
