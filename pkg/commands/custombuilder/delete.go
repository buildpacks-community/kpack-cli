package custombuilder

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewDeleteCommand(cmdContext commands.ContextProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete a custom builder",
		Long:    "Delete a custom builder from the provided namespace.\n If no namespace is provided, it attempts to delete the custo, builder from the default namespace",
		Example: "tbctl cb delete my-builder\ntbctl cb delete -n my-namespace other-builder",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := commands.InitContext(cmdContext, &namespace); err != nil {
				return err
			}

			err := cmdContext.KpackClient().ExperimentalV1alpha1().CustomBuilders(namespace).Delete(args[0], &metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" deleted\n", args[0])
			return err
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}
