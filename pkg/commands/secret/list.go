package secret

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/pivotal/build-service-cli/pkg/secret"
)

func NewListCommand(k8sClient k8s.Interface, defaultNamespace string) *cobra.Command {
	var namespace string

	command := cobra.Command{
		Use:          "list",
		Short:        "Display list of secrets",
		Long:         "Prints a table of the most important information about secrets. Only displays secrets in the current namespace.",
		Example:      "tbctl secret list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretList, err := k8sClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(secretList.Items) == 0 {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "no secrets found in %s namespace\n", namespace)
				return err
			} else {
				return displaySecretsTable(cmd, secretList)
			}
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "namespace name")

	return &command
}

func displaySecretsTable(cmd *cobra.Command, secretList *v1.SecretList) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 4, ' ', 0)

	_, err := fmt.Fprintln(writer, "NAME\tTARGET")
	if err != nil {
		return err
	}

	for _, item := range secretList.Items {
		var secretValue = item.Annotations[secret.RegistryAnnotation]
		if secretValue == "" {
			secretValue = item.Annotations[secret.GitAnnotation]
		}

		_, err := fmt.Fprintf(writer, "%s\t%s\n", item.Name, secretValue)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
