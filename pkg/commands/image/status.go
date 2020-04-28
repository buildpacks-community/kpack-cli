package image

import (
	"fmt"
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewStatusCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:          "status <name>",
		Short:        "Display image status",
		Long:         `Prints detailed information about the status of a specific image.`,
		Example:      "tbctl image status my-image\ntbctl image status my-other-image -n my-namespace",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			image, err := kpackClient.BuildV1alpha1().Images(namespace).Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			buildList, err := kpackClient.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			sort.Slice(buildList.Items, build.Sort(buildList.Items))
			return displayImageStatus(cmd, image, buildList.Items)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return cmd
}

func displayImageStatus(cmd *cobra.Command, image *v1alpha1.Image, builds []v1alpha1.Build) error {
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
		"Last Successful Build",
		"Id", getId(successfulBuild),
		"Reason", getReason(successfulBuild),
	)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"Last Failed Build",
		"Id", getId(failedBuild),
		"Reason", getReason(failedBuild),
	)
	if err != nil {
		return err
	}

	return statusWriter.Write()
}

func getLastSuccessfulBuild(builds []v1alpha1.Build) *v1alpha1.Build {
	for _, build := range builds {
		if build.IsSuccess() {
			return &build
		}
	}
	return nil
}

func getLastFailedBuild(builds []v1alpha1.Build) *v1alpha1.Build {
	for _, build := range builds {
		if build.IsFailure() {
			return &build
		}
	}
	return nil
}

type imageDetails struct {
	status      string
	message     string
	latestImage string
}

func getImageDetails(image *v1alpha1.Image) imageDetails {
	details := imageDetails{
		status:      "Unknown",
		message:     "",
		latestImage: "",
	}

	if cond := image.Status.GetCondition(v1alpha1.ConditionBuilderReady); cond != nil {
		if cond.Status != v1.ConditionTrue {
			details.status = "Not Ready"
			details.message = getNotReadyMessage(cond.Reason, image.Spec.Builder.Name)
			return details
		}
	}

	if cond := image.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == v1.ConditionTrue {
			details.status = "Ready"
		} else if cond.Status == v1.ConditionUnknown {
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
	if reason == v1alpha1.BuilderNotFound {
		return fmt.Sprintf("Builder '%s' not found", builderName)
	} else if reason == v1alpha1.BuilderNotReady {
		return fmt.Sprintf("Builder '%s' not ready", builderName)
	}
	return ""
}

func getId(build *v1alpha1.Build) string {
	if build == nil {
		return ""
	}
	if val, ok := build.Labels[v1alpha1.BuildNumberLabel]; ok {
		return val
	}
	return ""
}

func getReason(build *v1alpha1.Build) string {
	if build == nil {
		return ""
	}
	if val, ok := build.Annotations[v1alpha1.BuildReasonAnnotation]; ok {
		return val
	}
	return ""
}
