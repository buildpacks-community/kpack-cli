// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

func NewDeleteCommand(clientSetProvider k8s.ClientSetProvider, confirmationProvider ConfirmationProvider) *cobra.Command {
	const (
		warningMessage = "WARNING: Builders referring to buildpacks from this store will no longer schedule rebuilds for buildpack updates."
	)

	var (
		forceDelete bool
	)

	cmd := &cobra.Command{
		Use:          "delete <store>",
		Short:        "Delete a cluster store",
		Long:         fmt.Sprintf("Delete a specific cluster-scoped buildpack store.\n\n%s", warningMessage),
		Example:      `kp clusterstore delete my-store`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			storeName := args[0]
			if forceDelete {
				return deleteStore(cmd, cs, storeName)
			}

			message := fmt.Sprintf("%s\nPlease confirm store deletion by typing 'y': ", warningMessage)
			confirmed, err := confirmationProvider.Confirm(message)
			if err != nil {
				return err
			}

			if !confirmed {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "Skipping store deletion")
				return err
			}

			return deleteStore(cmd, cs, storeName)
		},
	}
	cmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func deleteStore(cmd *cobra.Command, cs k8s.ClientSet, storeName string) error {
	err := cs.KpackClient.ExperimentalV1alpha1().ClusterStores().Delete(storeName, &v1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return errors.Errorf("Store %q does not exist", storeName)
	} else if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" store deleted\n", storeName)
	return err
}
