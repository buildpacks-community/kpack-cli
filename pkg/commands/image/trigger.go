// Copyright 2020-Present VMware, Inc.
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

func NewTriggerCommand(clientSetProvider k8s.ClientSetProvider, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		namespace string
		wait      bool
	)

	cmd := &cobra.Command{
		Use:   "trigger <name>",
		Short: "Trigger an image build",
		Long: `Trigger a build using current inputs for a specific image in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: "kp image trigger my-image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			name := args[0]

			buildList, err := cs.KpackClient.KpackV1alpha1().Builds(cs.Namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + name,
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			}

			sort.Slice(buildList.Items, build.Sort(buildList.Items))
			bld := buildList.Items[len(buildList.Items)-1].DeepCopy()
			bld.Annotations[BuildNeededAnnotation] = time.Now().String()
			_, err = cs.KpackClient.KpackV1alpha1().Builds(cs.Namespace).Update(bld)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStderr(), "\"%s\" triggered\n", name)
			if err != nil {
				return err
			}

			if wait {
				img, err := cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Get(name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				_, err = newImageWaiter(cs).Wait(cmd.Context(), cmd.OutOrStdout(), img)
			}

			return err
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for image trigger to be reconciled and tail resulting build logs")

	return cmd
}
