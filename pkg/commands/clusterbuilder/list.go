// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available cluster builders",
		Long:         `Prints a table of the most important information about the available cluster builders.`,
		Example:      "kp clusterbuilder list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			clusterBuilderList, err := cs.KpackClient.KpackV1alpha2().ClusterBuilders().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(clusterBuilderList.Items) == 0 {
				return errors.New("no clusterbuilders found")
			} else {
				sort.Slice(clusterBuilderList.Items, Sort(clusterBuilderList.Items))
				return displayClusterBuildersTable(cmd, clusterBuilderList)
			}
		},
	}

	return cmd
}

func displayClusterBuildersTable(cmd *cobra.Command, builderList *v1alpha2.ClusterBuilderList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "Name", "Ready", "Stack", "Image")
	if err != nil {
		return err
	}

	for _, bldr := range builderList.Items {
		err := writer.AddRow(
			bldr.ObjectMeta.Name,
			getStatus(bldr),
			bldr.Status.Stack.ID,
			bldr.Status.LatestImage,
		)

		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func Sort(builds []v1alpha2.ClusterBuilder) func(i int, j int) bool {
	return func(i, j int) bool {
		return builds[j].ObjectMeta.Name > builds[i].ObjectMeta.Name
	}
}

func getStatus(b v1alpha2.ClusterBuilder) string {
	cond := b.Status.GetCondition(corev1alpha1.ConditionReady)
	switch {
	case cond.IsTrue():
		return "true"
	case cond.IsFalse():
		return "false"
	default:
		return "unknown"
	}
}
