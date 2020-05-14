package build

import (
	"sort"
	"strconv"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewStatusCommand(contextProvider commands.ContextProvider) *cobra.Command {
	var (
		namespace   string
		buildNumber int
	)

	cmd := &cobra.Command{
		Use:   "status <image-name>",
		Short: "Display image build status",
		Long: `Prints detailed information about the status of a specific image build.
If the build flag is not provided, the most recent build status will be shown.`,
		Example:      "tbctl build status my-image\ntbctl build status my-image -b 2 -n my-namespace",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := commands.GetContext(contextProvider, &namespace)
			if err != nil {
				return err
			}

			buildList, err := context.KpackClient.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))
				bld, err := findBuild(buildList, buildNumber, args[0], namespace)
				if err != nil {
					return err
				}
				return displayBuildStatus(cmd, bld)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().IntVarP(&buildNumber, "build", "b", -1, "build number")

	return cmd
}

func findBuild(buildList *v1alpha1.BuildList, buildNumber int, img, namespace string) (v1alpha1.Build, error) {
	if buildNumber == -1 {
		return buildList.Items[len(buildList.Items)-1], nil
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

	err := statusWriter.AddBlock(
		"",
		"Image", bld.Status.LatestImage,
		"Status", getStatus(bld),
		"Reasons", bld.Annotations[v1alpha1.BuildReasonAnnotation],
	)
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
