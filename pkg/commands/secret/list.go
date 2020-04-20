package secret

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/pivotal/build-service-cli/pkg/commands"
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
			serviceAccount, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get("default", metav1.GetOptions{})
			if err != nil {
				return err
			}

			if len(serviceAccount.Secrets) == 0 && len(serviceAccount.ImagePullSecrets) == 0 {
				return errors.Errorf("no secrets found in \"%s\" namespace", namespace)
			} else {
				return displaySecretsTable(cmd, serviceAccount)
			}
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return &command
}

func displaySecretsTable(cmd *cobra.Command, sa *v1.ServiceAccount) error {
	managedSecrets, err := readManagedSecrets(sa)
	if err != nil {
		return err
	}

	secretNameSet := map[string]interface{}{}
	for _, item := range append(sa.Secrets) {
		secretNameSet[item.Name] = nil
	}
	for _, item := range append(sa.ImagePullSecrets) {
		secretNameSet[item.Name] = nil
	}

	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "TARGET")
	if err != nil {
		return err
	}

	for name := range secretNameSet {
		err := writer.AddRow(name, managedSecrets[name])
		if err != nil {
			return err
		}
	}

	return writer.Write()
}
