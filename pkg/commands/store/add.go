// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/store"
)

func NewAddCommand(clientSetProvider k8s.ClientSetProvider, factory *store.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <store> <buildpackage> [<buildpackage>...]",
		Short: "Add buildpackage(s) to store",
		Long: `Upload buildpackage(s) to a specific buildpack store.

Buildpackages will be uploaded to the the registry configured on your store.
Therefore, you must have credentials to access the registry on your machine.`,
		Example: `kp store add my-store my-registry.com/my-buildpackage
kp store add my-store my-registry.com/my-buildpackage my-registry.com/my-other-buildpackage my-registry.com/my-third-buildpackage
kp store add my-store ../path/to/my-local-buildpackage.cnb`,
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			storeName := args[0]
			buildpackages := args[1:]
			factory.Printer = commands.NewPrinter(cmd)

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			s, err := cs.KpackClient.ExperimentalV1alpha1().Stores().Get(storeName, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("Store '%s' does not exist", storeName)
			} else if err != nil {
				return err
			}

			updatedStore, storeUpdated, err := factory.AddToStore(s, buildpackages...)
			if err != nil {
				return err
			}

			if !storeUpdated {
				factory.Printer.Printf("Store Unchanged")
				return nil
			}

			_, err = cs.KpackClient.ExperimentalV1alpha1().Stores().Update(updatedStore)
			if err != nil {
				return err
			}

			factory.Printer.Printf("Store Updated")
			return nil
		},
	}
	return cmd
}
