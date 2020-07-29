// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewAddCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <store> <buildpackage> [<buildpackage>...]",
		Short: "Add buildpackage(s) to cluster store",
		Long: `Upload buildpackage(s) to a specific cluster-scoped buildpack store.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore add my-store my-registry.com/my-buildpackage
kp clusterstore add my-store my-registry.com/my-buildpackage --buildpackage my-registry.com/my-other-buildpackage --buildpackage my-registry.com/my-third-buildpackage
kp clusterstore add my-store --buildpackage ../path/to/my-local-buildpackage.cnb`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			storeName := args[0]
			factory.Printer = commands.NewPrinter(cmd)

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			s, err := cs.KpackClient.ExperimentalV1alpha1().ClusterStores().Get(storeName, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("Store '%s' does not exist", storeName)
			} else if err != nil {
				return err
			}

			repo, err := k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
			if err != nil {
				return err
			}

			updatedStore, storeUpdated, err := factory.AddToStore(s, repo, buildpackages...)
			if err != nil {
				return err
			}

			if !storeUpdated {
				factory.Printer.Printf("ClusterStore Unchanged")
				return nil
			}

			_, err = cs.KpackClient.ExperimentalV1alpha1().ClusterStores().Update(updatedStore)
			if err != nil {
				return err
			}

			factory.Printer.Printf("ClusterStore Updated")
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	return cmd
}
