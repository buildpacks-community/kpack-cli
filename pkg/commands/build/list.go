package build

import (
	"sort"
	"strings"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewListCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "list <name>",
		Short: "List image builds",
		Long: `Prints a table of the most important information about an image's builds.
Will only display builds in your current namespace.
If no namespace is provided, the default namespace is queried.`,
		Example:      "tbctl image build list\ntbctl image build list -n my-namespace",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			buildList, err := kpackClient.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.Errorf("no builds for image \"%s\" found in \"%s\" namespace", args[0], namespace)
			} else {
				sort.Slice(buildList.Items, sortBuilds(buildList.Items))
				return displayBuildsTable(cmd, buildList)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return cmd
}

func displayBuildsTable(cmd *cobra.Command, buildList *v1alpha1.BuildList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "Build", "Status", "Image", "Started", "Finished", "Reason")
	if err != nil {
		return err
	}

	for _, bld := range buildList.Items {
		err := writer.AddRow(
			bld.Labels[v1alpha1.BuildNumberLabel],
			getStatus(bld),
			bld.Status.LatestImage,
			getStarted(bld),
			getFinished(bld),
			getTruncatedReason(bld),
		)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func sortBuilds(builds []v1alpha1.Build) func(i int, j int) bool {
	return func(i, j int) bool {
		return builds[j].ObjectMeta.CreationTimestamp.After(builds[i].ObjectMeta.CreationTimestamp.Time)
	}
}

func getStatus(b v1alpha1.Build) string {
	cond := b.Status.GetCondition(corev1alpha1.ConditionSucceeded)
	switch {
	case cond.IsTrue():
		return "SUCCESS"
	case cond.IsFalse():
		return "FAILURE"
	case cond.IsUnknown():
		return "BUILDING"
	default:
		return "UNKNOWN"
	}
}

func getStarted(b v1alpha1.Build) string {
	return b.CreationTimestamp.Time.Format("2006-01-02 15:04:05")
}

func getFinished(b v1alpha1.Build) string {
	if b.IsRunning() {
		return ""
	}
	return b.Status.GetCondition(corev1alpha1.ConditionSucceeded).LastTransitionTime.Inner.Format("2006-01-02 15:04:05")
}

func getTruncatedReason(b v1alpha1.Build) string {
	r := getReasons(b)

	if len(r) == 0 {
		return "UNKNOWN"
	}

	if len(r) == 1 {
		return r[0]
	}

	return mostImportantReason(r) + "+"
}

func getReasons(b v1alpha1.Build) []string {
	s := strings.Split(b.Annotations[v1alpha1.BuildReasonAnnotation], ",")
	if len(s) == 1 && s[0] == "" {
		return nil
	}
	return s
}

func mostImportantReason(r []string) string {
	if contains(r, "CONFIG") {
		return "CONFIG"
	} else if contains(r, "COMMIT") {
		return "COMMIT"
	} else if contains(r, "BUILDPACK") {
		return "BUILDPACK"
	}

	return r[0]
}

func contains(reasons []string, value string) bool {
	for _, v := range reasons {
		if v == value {
			return true
		}
	}
	return false
}
