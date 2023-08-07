// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/build"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
		allNamespaces bool
	)

	cmd := &cobra.Command{
		Use:   "list [image-resource-name]",
		Short: "List builds",
		Long: `Prints a table of the most important information about builds in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,

		Example:      "kp build list\nkp build list my-image\nkp build list my-image -n my-namespace\nkp build list -A",
		Args:         commands.OptionalArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			opts := metav1.ListOptions{}

			if len(args) > 0 {
				opts.LabelSelector = v1alpha2.ImageLabel + "=" + args[0]
			}

			var buildNamespace string

			if allNamespaces {
				buildNamespace = ""
			} else {
				buildNamespace = cs.Namespace
			}

			buildList, err := cs.KpackClient.KpackV1alpha2().Builds(buildNamespace).List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))
				return displayBuildsTable(cmd, buildList)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Return objects found in all namespaces")

	return cmd
}

func displayBuildsTable(cmd *cobra.Command, buildList *v1alpha2.BuildList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "Build", "Status", "Built Image", "Reason", "Image Resource")
	if err != nil {
		return err
	}

	for _, bld := range buildList.Items {
		err := writer.AddRow(
			bld.Labels[v1alpha2.BuildNumberLabel],
			getStatus(bld),
			bld.Status.LatestImage,
			getTruncatedReason(bld),
			bld.Labels[v1alpha2.ImageLabel],
		)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}
