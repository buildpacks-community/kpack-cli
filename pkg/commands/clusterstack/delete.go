// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewDeleteCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete a cluster stack",
		Long:    "Delete a specific cluster-scoped stack from the cluster.",
		Example: "kp clusterstack delete my-stack",
		Args:    commands.ExactArgsWithUsage(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			err = cs.KpackClient.KpackV1alpha2().ClusterStacks().Delete(cmd.Context(), args[0], metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "ClusterStack %q deleted\n", args[0])
			return err
		},
		SilenceUsage: true,
	}

	return cmd
}
