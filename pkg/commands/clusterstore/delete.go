// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
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

			ctx := cmd.Context()

			storeName := args[0]
			if forceDelete {
				return deleteStore(ctx, cmd, cs, storeName)
			}

			message := fmt.Sprintf("%s\nPlease confirm store deletion by typing 'y': ", warningMessage)
			confirmed, err := confirmationProvider.Confirm(message)
			if err != nil {
				return err
			}

			if !confirmed {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "Skipping ClusterStore deletion")
				return err
			}

			return deleteStore(ctx, cmd, cs, storeName)
		},
	}
	cmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func deleteStore(ctx context.Context, cmd *cobra.Command, cs k8s.ClientSet, storeName string) error {
	err := cs.KpackClient.KpackV1alpha2().ClusterStores().Delete(ctx, storeName, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return errors.Errorf("Store %q does not exist", storeName)
	} else if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "ClusterStore %q store deleted\n", storeName)
	return err
}
