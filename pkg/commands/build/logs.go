// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/logs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewLogsCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace   string
		buildNumber string
	)

	cmd := &cobra.Command{
		Use:   "logs <image-name>",
		Short: "Tails logs for an image build",
		Long: `Tails logs from the containers of a specific build of an image in the provided namespace.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp build logs my-image\nkp build logs my-image -b 2 -n my-namespace",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			buildList, err := cs.KpackClient.KpackV1alpha1().Builds(cs.Namespace).List(cmd.Context(), metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))
				bld, err := findBuild(buildList, buildNumber)
				if err != nil {
					return err
				}
				return logs.NewBuildLogsClient(cs.K8sClient).Tail(context.Background(), cmd.OutOrStdout(), args[0], bld.Labels[v1alpha1.BuildNumberLabel], cs.Namespace)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&buildNumber, "build", "b", "", "build number")

	return cmd
}
