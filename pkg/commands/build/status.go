// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"sort"
	"strconv"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace   string
		buildNumber string
	)

	cmd := &cobra.Command{
		Use:   "status <image-name>",
		Short: "Display status for an image build",
		Long: `Prints detailed information about the status of a specific build of an image in the provided namespace.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp build status my-image\nkp build status my-image -b 2 -n my-namespace",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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
				bld, err := findBuild(buildList, buildNumber, args[0], cs.Namespace)
				if err != nil {
					return err
				}
				return displayBuildStatus(cmd, bld)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&buildNumber, "build", "b", "", "build number")

	return cmd
}

func findBuild(buildList *v1alpha1.BuildList, buildNumberString string, img, namespace string) (v1alpha1.Build, error) {

	if buildNumberString == "" {
		return buildList.Items[len(buildList.Items)-1], nil
	}

	buildNumber, err := strconv.Atoi(buildNumberString)
	if err != nil {
		return v1alpha1.Build{}, errors.Errorf("build number should be an integer: %v", buildNumberString)
	}

	for _, b := range buildList.Items {
		val, err := strconv.Atoi(b.Labels[v1alpha1.BuildNumberLabel])
		if err != nil {
			return v1alpha1.Build{}, err
		}

		if val == buildNumber {
			return b, nil
		}
	}

	return v1alpha1.Build{}, errors.Errorf("build \"%d\" not found", buildNumber)
}

func displayBuildStatus(cmd *cobra.Command, bld v1alpha1.Build) error {
	statusWriter := commands.NewStatusWriter(cmd.OutOrStdout())

	statusItems := []string{
		"Image", bld.Status.LatestImage,
		"Status", getStatus(bld),
		"Build Reasons", bld.Annotations[v1alpha1.BuildReasonAnnotation],
	}

	if cond := bld.Status.GetCondition(corev1alpha1.ConditionSucceeded); cond.Reason != "" {
		statusItems = append(statusItems, "Status Reason", cond.Reason)
	}
	if cond := bld.Status.GetCondition(corev1alpha1.ConditionSucceeded); cond.Message != "" {
		statusItems = append(statusItems, "Status Message", cond.Message)
	}

	err := statusWriter.AddBlock("", statusItems...)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock("",
		"Pod Name", bld.Status.PodName)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"",
		"Builder", bld.Spec.Builder.Image,
		"Run Image", bld.Status.Stack.RunImage,
	)
	if err != nil {
		return err
	}

	if bld.Spec.Source.Git != nil {
		err = statusWriter.AddBlock(
			"",
			"Source", "Git",
			"Url", bld.Spec.Source.Git.URL,
			"Revision", bld.Spec.Source.Git.Revision,
		)
		if err != nil {
			return err
		}
	} else if bld.Spec.Source.Blob != nil {
		err = statusWriter.AddBlock(
			"",
			"Source", "Blob",
			"Url", bld.Spec.Source.Blob.URL,
		)
		if err != nil {
			return err
		}
	} else {
		err = statusWriter.AddBlock("", "Source", "Local Source")
		if err != nil {
			return err
		}
	}

	err = statusWriter.Write()
	if err != nil {
		return err
	}

	tableWriter, err := commands.NewTableWriter(cmd.OutOrStdout(), "Buildpack Id", "Buildpack Version")
	if err != nil {
		return err
	}

	for _, buildpack := range bld.Status.BuildMetadata {
		err := tableWriter.AddRow(buildpack.Id, buildpack.Version)
		if err != nil {
			return err
		}
	}

	return tableWriter.Write()
}
