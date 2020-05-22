package store

import (
	"errors"

	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images",
		Long: `Prints a table of the most important information about images in the provided namespace.

namespace defaults to the kubernetes current-context namespace.`,
		Example: "tbctl image list\ntbctl image list -n my-namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			storeList, err := cs.KpackClient.ExperimentalV1alpha1().Stores().List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(storeList.Items) == 0 {
				return errors.New("no stores found")
			} else {
				return displayStoresTable(cmd, storeList)
			}

		},
		SilenceUsage: true,
	}

	return cmd
}

func displayStoresTable(cmd *cobra.Command, storeList *expv1alpha1.StoreList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY")
	if err != nil {
		return err
	}

	for _, s := range storeList.Items {
		err := writer.AddRow(s.Name, getReadyText(s))
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getReadyText(s expv1alpha1.Store) string {
	cond := s.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
