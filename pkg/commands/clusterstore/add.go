// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewAddCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	var (
		buildpackages []string
		dryRunConfig  DryRunConfig
	)

	cmd := &cobra.Command{
		Use:   "add <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Add buildpackage(s) to cluster store",
		Long: `Upload buildpackage(s) to a specific cluster-scoped buildpack store.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore add my-store -b my-registry.com/my-buildpackage
kp clusterstore add my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage -b my-registry.com/my-third-buildpackage
kp clusterstore add my-store -b ../path/to/my-local-buildpackage.cnb`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			storeName := args[0]
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

			s, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(storeName, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("ClusterStore '%s' does not exist", storeName)
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

func update(s *v1alpha1.ClusterStore, buildpackages []string, factory *clusterstore.Factory, drc DryRunConfig, cs k8s.ClientSet) error {
	repo, err := k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	factory.Printer.Printf("Adding Buildpackages...")
	updatedStore, storeUpdated, err := factory.AddToStore(s, repo, buildpackages...)
	if err != nil {
		return err
	}

	if !storeUpdated {
		factory.Printer.Printf("ClusterStore Unchanged")
		return nil
	}

	if drc.dryRun {
		printer, err := commands.NewResourcePrinter(drc.outputFormat)
		if err != nil {
			return err
		}

		return printer.PrintObject(updatedStore, drc.writer)
	}

	_, err = cs.KpackClient.KpackV1alpha1().ClusterStores().Update(updatedStore)
	if err != nil {
		return err
	}

	factory.Printer.Printf("ClusterStore Updated")
	return nil
}
