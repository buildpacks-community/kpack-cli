// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/logs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/build"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewLogsCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace   string
		buildNumber string
	)

	cmd := &cobra.Command{
		Use:   "logs <image-name|build-name>",
		Short: "Tails logs for an image resource or for a build resource",
		Long: `Tails logs from the containers of a specific build of an image or build resource in the provided namespace.

By default command will assume user provided an Image name and will attempt to find builds associated with that Image.
If no builds are found matching the Image name, It will assume the provided argument was a Build name.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.

Use the flag --timestamps to include the timestamps for the logs`,
		Example: `kp build logs my-image
kp build logs my-image -b 2 -n my-namespace
kp build logs my-build-name -n my-namespace`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			// find all builds associated with that are created by an image with the provided name
			buildList, err := cs.KpackClient.KpackV1alpha2().Builds(cs.Namespace).List(cmd.Context(), metav1.ListOptions{
				LabelSelector: v1alpha2.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			// no image exists, lets see if a build exists with this name
			if len(buildList.Items) == 0 {
				build, err := cs.KpackClient.KpackV1alpha2().Builds(cs.Namespace).Get(cmd.Context(), args[0], metav1.GetOptions{})
				if err != nil {
					// no build exists and no image exists, so we return error
					return errors.New("no builds found")
				}
				// build found, tail logs
				return logs.NewBuildLogsClient(cs.K8sClient).TailBuildName(context.Background(), cmd.OutOrStdout(), cs.Namespace, build.Name, ch.ShowTimestamp())

			}

			// an image exists with the provided name, tail build logs (default to latest, or use -b for a specific build)
			sort.Slice(buildList.Items, build.Sort(buildList.Items))
			bld, err := findBuild(buildList, buildNumber)
			if err != nil {
				return err
			}
			return logs.NewBuildLogsClient(cs.K8sClient).Tail(context.Background(), cmd.OutOrStdout(), args[0], bld.Labels[v1alpha2.BuildNumberLabel], cs.Namespace, ch.ShowTimestamp())

		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&buildNumber, "build", "b", "", "build number")
	cmd.Flags().BoolP("timestamps", "t", false, "show log timestamps")
	return cmd
}
