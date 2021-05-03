// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List cluster stacks",
		Long:         `Prints a table of the most important information about cluster-scoped stacks in the cluster.`,
		Example:      "kp clusterstack list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			stackList, err := cs.KpackClient.KpackV1alpha1().ClusterStacks().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(stackList.Items) == 0 {
				return errors.New("no clusterstacks found")
			} else {
				return displayStacksTable(cmd, stackList)
			}

		},
	}

	return cmd
}

func displayStacksTable(cmd *cobra.Command, stackList *v1alpha1.ClusterStackList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY", "ID")
	if err != nil {
		return err
	}

	for _, s := range stackList.Items {
		err := writer.AddRow(s.Name, getReadyText(s), s.Status.Id)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getReadyText(s v1alpha1.ClusterStack) string {
	cond := s.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
