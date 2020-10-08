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
		dryRun        bool
		output        string
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

			s, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(name, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("ClusterStore '%s' does not exist", name)
			} else if err != nil {
				return err
			}

			return update(s, buildpackages, factory, ch, cs)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&output, "output", "", "output format. supported formats are: yaml, json")
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	return cmd
}

func update(store *v1alpha1.ClusterStore, buildpackages []string, factory *clusterstore.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) error {
	repo, err := k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	if err = ch.PrintStatus("Adding Buildpackages..."); err != nil {
		return err
	}

	updatedStore, storeUpdated, err := factory.AddToStore(store, repo, buildpackages...)
	if err != nil {
		return err
	}

	if storeUpdated && !ch.IsDryRun() {
		updatedStore, err = cs.KpackClient.KpackV1alpha1().ClusterStores().Update(updatedStore)
		if err != nil {
			return err
		}
	}

	if err = ch.PrintObj(updatedStore); err != nil {
		return err
	}

	return ch.PrintChangeResult(storeUpdated, "ClusterStore Updated")
}
