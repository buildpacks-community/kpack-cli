// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List cluster lifecycles",
		Long:         `Prints a table of the most important information about cluster-scoped lifecycles in the cluster.`,
		Example:      "kp clusterlifecycle list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			lifecycleList, err := cs.KpackClient.KpackV1alpha2().ClusterLifecycles().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(lifecycleList.Items) == 0 {
				return errors.New("no clusterlifecycles found")
			} else {
				return displayLifecyclesTable(cmd, lifecycleList)
			}

		},
	}

	return cmd
}

func displayLifecyclesTable(cmd *cobra.Command, lifecycleList *v1alpha2.ClusterLifecycleList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY", "VERSION", "IMAGE")
	if err != nil {
		return err
	}

	for _, l := range lifecycleList.Items {
		err := writer.AddRow(l.Name, getReadyText(l), l.Status.Version, l.Status.Image.LatestImage)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getReadyText(s v1alpha2.ClusterLifecycle) string {
	cond := s.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
