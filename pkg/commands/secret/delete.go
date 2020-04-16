package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

func NewDeleteCommand(k8sClient k8s.Interface, defaultNamespace string) *cobra.Command {
	var namespace string

	command := cobra.Command{
		Use:          "delete <name>",
		Short:        "Delete secret",
		Long:         "Deletes the provided secret from the desired namespace.",
		Example:      "tbctl secret delete my-secret",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := k8sClient.CoreV1().Secrets(namespace).Delete(args[0], &metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" deleted\n", args[0])
			return err
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return &command
}
