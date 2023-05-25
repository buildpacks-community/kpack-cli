// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack

import (
	"sort"

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
		Use:   "list",
		Short: "List available cluster buildpacks",
		Long: `Prints a table of the most important information about the available cluster buildpacks in the provided namespace.
`,
		Example:      "kp clusterbuildpack list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			cbpList, err := cs.KpackClient.KpackV1alpha2().ClusterBuildpacks().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(cbpList.Items) == 0 {
				return errors.New("no cluster buildpacks found")
			} else {
				sort.Slice(cbpList.Items, Sort(cbpList.Items))

				return displayClusterBuildpacksTable(cmd, cbpList)
			}
		},
	}

	return cmd
}

func displayClusterBuildpacksTable(cmd *cobra.Command, cbpList *v1alpha2.ClusterBuildpackList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "Name", "Ready", "Image")
	if err != nil {
		return err
	}

	for _, cbp := range cbpList.Items {
		err := writer.AddRow(
			cbp.Name,
			getStatus(cbp),
			cbp.Spec.Image,
		)

		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func Sort(cbps []v1alpha2.ClusterBuildpack) func(i int, j int) bool {
	return func(i, j int) bool {
		return cbps[j].Name > cbps[i].Name
	}
}

func getStatus(b v1alpha2.ClusterBuildpack) string {
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
