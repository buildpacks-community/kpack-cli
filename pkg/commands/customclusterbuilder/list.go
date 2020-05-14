package customclusterbuilder

import (
	"sort"

	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewListCommand(cmdContext commands.ContextProvider) *cobra.Command {

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available custom cluster builders",
		Long:         `Prints a table of the most important information about the available custom cluster builders.`,
		Example:      "tbctl ccb list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdContext.Initialize(); err != nil {
				return err
			}

			clusterBuilderList, err := cmdContext.KpackClient().ExperimentalV1alpha1().CustomClusterBuilders().List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(clusterBuilderList.Items) == 0 {
				return errors.New("no clusterbuilders found")
			} else {
				sort.Slice(clusterBuilderList.Items, Sort(clusterBuilderList.Items))
				return displayClusterBuildersTable(cmd, clusterBuilderList)
			}
		},
	}

	return cmd
}

func displayClusterBuildersTable(cmd *cobra.Command, builderList *expv1alpha1.CustomClusterBuilderList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "Name", "Ready", "Stack", "Image")
	if err != nil {
		return err
	}

	for _, bldr := range builderList.Items {
		err := writer.AddRow(
			bldr.ObjectMeta.Name,
			getStatus(bldr),
			bldr.Status.Stack.ID,
			bldr.Status.LatestImage,
		)

		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func Sort(builds []expv1alpha1.CustomClusterBuilder) func(i int, j int) bool {
	return func(i, j int) bool {
		return builds[j].ObjectMeta.Name > builds[i].ObjectMeta.Name
	}
}

func getStatus(b expv1alpha1.CustomClusterBuilder) string {
	cond := b.Status.GetCondition(corev1alpha1.ConditionReady)
	switch {
	case cond.IsTrue():
		return "true"
	case cond.IsFalse():
		return "false"
	default:
		return "unknown"
	}
}
