package build

import (
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewListCommand(contextProvider commands.ContextProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "list <image-name>",
		Short: "List image builds",
		Long: `Prints a table of the most important information about an image's builds.
Will only display builds in your current namespace.
If no namespace is provided, the default namespace is queried.`,
		Example:      "tbctl build list my-image\ntbctl build list my-image -n my-namespace",
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
				return displayBuildsTable(cmd, buildList)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

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
