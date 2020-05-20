package stack

import (
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List stacks",
		Long:         `Prints a table of the most important information about stacks in the cluster.`,
		Example:      "tbctl stack list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			stackList, err := cs.KpackClient.ExperimentalV1alpha1().Stacks().List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(stackList.Items) == 0 {
				return errors.New("no stacks found")
			} else {
				return displayStacksTable(cmd, stackList)
			}

		},
	}

	return cmd
}

func displayStacksTable(cmd *cobra.Command, stackList *expv1alpha1.StackList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY", "ID")
	if err != nil {
		return err
	}

	for _, s := range stackList.Items {
		err := writer.AddRow(s.Name, getReadyText(s), s.Status.Id)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getReadyText(s expv1alpha1.Stack) string {
	cond := s.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
