// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/kpack-cli/pkg/build"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

const BuildNeededAnnotation = "image.kpack.io/additionalBuildNeeded"

func NewTriggerCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "trigger <name>",
		Short: "Trigger an image resource build",
		Long: `Trigger a build using current inputs for a specific image resource in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: "kp image trigger my-image",
		Args:    commands.ExactArgsWithUsage(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			buildList, err := cs.KpackClient.KpackV1alpha2().Builds(cs.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: v1alpha2.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))

				original := buildList.Items[len(buildList.Items)-1].DeepCopy()
				patched := original.DeepCopy()
				patched.Annotations[BuildNeededAnnotation] = time.Now().String()

				patch, err := k8s.CreatePatch(original, patched)
				if err != nil {
					return err
				}

				_, err = cs.KpackClient.KpackV1alpha2().Builds(cs.Namespace).Patch(ctx, original.Name, types.MergePatchType, patch, metav1.PatchOptions{})
				if err != nil {
					return err
				}

				previousBuildNumber, _ := strconv.Atoi(patched.Labels[v1alpha2.BuildNumberLabel])

				_, err = fmt.Fprintf(cmd.OutOrStderr(), "Triggered build for Image Resource %q with Build Number %d\n", args[0], previousBuildNumber+1)
				return err
			}
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}
