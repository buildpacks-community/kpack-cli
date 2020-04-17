package secret

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/secret"
)

func NewListCommand(k8sClient k8s.Interface, defaultNamespace string) *cobra.Command {
	var namespace string

	command := cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Long: `Prints a table of the most important information about secrets.
Will only display secrets in the current namespace.
If no namespace is provided, the default namespace is queried.`,
		Example:      "tbctl secret list\ntbctl secret list -n my-namespace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretList, err := k8sClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(secretList.Items) == 0 {
				return errors.Errorf("no secrets found in \"%s\" namespace", namespace)
			} else {
				return displaySecretsTable(cmd, secretList)
			}
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return &command
}

func displaySecretsTable(cmd *cobra.Command, secretList *v1.SecretList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "TARGET")
	if err != nil {
		return err
	}

	for _, item := range secretList.Items {
		var secretValue = item.Annotations[secret.TargetAnnotation]
		if secretValue == "" {
			secretValue = item.Annotations[secret.GitAnnotation]
		}

		err := writer.AddRow(item.Name, secretValue)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}
