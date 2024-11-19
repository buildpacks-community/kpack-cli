// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack

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
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available buildpacks",
		Long: `Prints a table of the most important information about the available buildpacks in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp buildpack list\nkp buildpack list -n my-namespace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			bpList, err := cs.KpackClient.KpackV1alpha2().Buildpacks(cs.Namespace).List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(bpList.Items) == 0 {
				return errors.New("no buildpacks found")
			} else {
				sort.Slice(bpList.Items, Sort(bpList.Items))

				return displayBuildpacksTable(cmd, bpList)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}

func displayBuildpacksTable(cmd *cobra.Command, bpList *v1alpha2.BuildpackList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "Name", "Ready", "Image")
	if err != nil {
		return err
	}

	for _, bp := range bpList.Items {
		err := writer.AddRow(
			bp.Name,
			getStatus(bp),
			bp.Spec.Image,
		)

		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func Sort(bps []v1alpha2.Buildpack) func(i int, j int) bool {
	return func(i, j int) bool {
		return bps[j].Name > bps[i].Name
	}
}

func getStatus(b v1alpha2.Buildpack) string {
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
