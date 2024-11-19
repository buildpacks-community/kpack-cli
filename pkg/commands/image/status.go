// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/build"
	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Display status of an image resource",
		Long: `Prints detailed information about the status of a specific image resource in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp image status my-image\nkp image status my-other-image -n my-namespace",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			image, err := cs.KpackClient.KpackV1alpha2().Images(cs.Namespace).Get(ctx, args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			buildList, err := cs.KpackClient.KpackV1alpha2().Builds(cs.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: v1alpha2.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			sort.Slice(buildList.Items, build.Sort(buildList.Items))
			return displayImageStatus(cmd, image, buildList.Items)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}

func displayImageStatus(cmd *cobra.Command, image *v1alpha2.Image, builds []v1alpha2.Build) error {
	statusWriter := commands.NewStatusWriter(cmd.OutOrStdout())
	imgDetails := getImageDetails(image)
	failedBuild := getLastFailedBuild(builds)
	successfulBuild := getLastSuccessfulBuild(builds)

	err := statusWriter.AddBlock(
		"",
		"Status", imgDetails.status,
		"Message", imgDetails.message,
		"LatestImage", imgDetails.latestImage,
	)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"Source",
		getConfigSource(image)...,
	)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"Builder Ref",
		"Name", image.Spec.Builder.Name,
		"Kind", image.Spec.Builder.Kind,
	)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"Last Successful Build",
		buildStatus(successfulBuild)...,
	)
	if err != nil {
		return err
	}

	if successfulBuild != nil {
		tableWriter, err := commands.NewTableWriter(cmd.OutOrStdout(), "Buildpack Id", "Buildpack Version", "Homepage")
		if err != nil {
			return err
		}

		for _, metadata := range successfulBuild.Status.BuildMetadata {
			err := tableWriter.AddRow(metadata.Id, metadata.Version, metadata.Homepage)
			if err != nil {
				return err
			}
		}
		err = tableWriter.Write()
		if err != nil {
			return err
		}
	}

	err = statusWriter.AddBlock(
		"Last Failed Build",
		buildStatus(failedBuild)...,
	)
	if err != nil {
		return err
	}

	return statusWriter.Write()
}

func buildStatus(build *v1alpha2.Build) []string {
	items := []string{
		"Id", getId(build),
		"Build Reason", getReason(build),
	}
	if build != nil && build.Spec.Source.Git != nil {
		items = append(items, "Git Revision", build.Spec.Source.Git.Revision)
	}
	return items
}

func getConfigSource(image *v1alpha2.Image) []string {
	if image.Spec.Source.Git != nil {
		return []string{
			"Type", "GitUrl",
			"Url", image.Spec.Source.Git.URL,
			"Revision", image.Spec.Source.Git.Revision,
		}
	} else if image.Spec.Source.Blob != nil {
		return []string{
			"Type", "Blob",
			"Url", image.Spec.Source.Blob.URL,
		}
	} else {
		return []string{
			"Type", "Local Source"}
	}
}

func getLastSuccessfulBuild(builds []v1alpha2.Build) *v1alpha2.Build {
	for i, _ := range builds {
		if builds[len(builds)-1-i].IsSuccess() {
			return &builds[len(builds)-1-i]
		}
	}
	return nil
}

func getLastFailedBuild(builds []v1alpha2.Build) *v1alpha2.Build {
	for i, _ := range builds {
		if builds[len(builds)-1-i].IsFailure() {
			return &builds[len(builds)-1-i]
		}
	}
	return nil
}

type imageDetails struct {
	status      string
	message     string
	latestImage string
}

func getImageDetails(image *v1alpha2.Image) imageDetails {
	details := imageDetails{
		status:      "Unknown",
		message:     "",
		latestImage: "",
	}

	if cond := image.Status.GetCondition(v1alpha2.ConditionBuilderReady); cond != nil {
		if cond.Status != corev1.ConditionTrue {
			details.status = "Not Ready"
			details.message = getNotReadyMessage(cond.Reason, image.Spec.Builder.Name)
			return details
		}
	}

	if cond := image.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == corev1.ConditionTrue {
			details.status = "Ready"
		} else if cond.Status == corev1.ConditionUnknown {
			details.status = "Building"
		} else {
			details.status = "Not Ready"
			details.message = getNotReadyMessage(cond.Reason, image.Spec.Builder.Name)
		}
	}

	if image.Status.LatestImage != "" {
		details.latestImage = image.Status.LatestImage
	}

	return details
}

func getNotReadyMessage(reason, builderName string) string {
	if reason == v1alpha2.BuilderNotFound {
		return fmt.Sprintf("Builder '%s' not found", builderName)
	} else if reason == v1alpha2.BuilderNotReady {
		return fmt.Sprintf("Builder '%s' not ready", builderName)
	}
	return ""
}

func getId(build *v1alpha2.Build) string {
	if build == nil {
		return ""
	}
	if val, ok := build.Labels[v1alpha2.BuildNumberLabel]; ok {
		return val
	}
	return ""
}

func getReason(build *v1alpha2.Build) string {
	if build == nil {
		return ""
	}
	if val, ok := build.Annotations[v1alpha2.BuildReasonAnnotation]; ok {
		return val
	}
	return ""
}
