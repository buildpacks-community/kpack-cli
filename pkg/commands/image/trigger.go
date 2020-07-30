// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"sort"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const BuildNeededAnnotation = "image.kpack.io/additionalBuildNeeded"

func NewTriggerCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "trigger <name>",
		Short: "Trigger an image build",
		Long: `Trigger a build using current inputs for a specific image in the provided namespace.

namespace defaults to the kubernetes current-context namespace.`,
		Example: "kp image trigger my-image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			buildList, err := cs.KpackClient.KpackV1alpha1().Builds(cs.Namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))

				build := buildList.Items[len(buildList.Items)-1].DeepCopy()
				build.Annotations[BuildNeededAnnotation] = time.Now().String()
				_, err := cs.KpackClient.KpackV1alpha1().Builds(cs.Namespace).Update(build)
				if err != nil {
					return err
				}

				_, err = fmt.Fprintf(cmd.OutOrStderr(), "\"%s\" triggered\n", args[0])
				return err
			}
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}
