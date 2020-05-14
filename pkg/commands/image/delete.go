package image

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
		Short:   "Delete an image",
		Long:    "Delete an image and the associated image builds from the cluster.",
		Example: "tbctl image delete my-image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := commands.InitContext(cmdContext, &namespace); err != nil {
				return err
			}

			err := cmdContext.KpackClient().BuildV1alpha1().Images(namespace).Delete(args[0], &metav1.DeleteOptions{})
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
