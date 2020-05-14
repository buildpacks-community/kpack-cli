package stack

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewDeleteCommand(contextProvider commands.ContextProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete a stack",
		Long:    "Delete a stack from the cluster.",
		Example: "tbctl stack delete my-stack",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := contextProvider.GetContext()
			if err != nil {
				return err
			}

			err = context.KpackClient.ExperimentalV1alpha1().Stacks().Delete(args[0], &metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" deleted\n", args[0])
			return err
		},
		SilenceUsage: true,
	}

	return cmd
}
