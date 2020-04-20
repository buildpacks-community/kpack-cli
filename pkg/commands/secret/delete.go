package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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
			serviceAccount, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get("default", metav1.GetOptions{})
			if err != nil {
				return err
			}

			wasModified, err := deleteSecretsFromServiceAccount(serviceAccount, args[0])
			if err != nil {
				return err
			} else if wasModified {
				_, err = k8sClient.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
				if err != nil {
					return err
				}
			}

			err = k8sClient.CoreV1().Secrets(namespace).Delete(args[0], &metav1.DeleteOptions{})
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

func deleteSecretsFromServiceAccount(sa *corev1.ServiceAccount, name string) (bool, error) {
	managedSecrets, err := readManagedSecrets(sa)
	if err != nil {
		return false, err
	}

	modified := false
	for i, s := range sa.Secrets {
		if s.Name == name {
			sa.Secrets = append(sa.Secrets[:i], sa.Secrets[i+1:]...)
			delete(managedSecrets, s.Name)
			modified = true
			break
		}
	}
	for i, s := range sa.ImagePullSecrets {
		if s.Name == name {
			sa.ImagePullSecrets = append(sa.ImagePullSecrets[:i], sa.ImagePullSecrets[i+1:]...)
			delete(managedSecrets, s.Name)
			modified = true
			break
		}
	}

	err = writeManagedSecrets(managedSecrets, sa)
	if err != nil {
		return false, err
	}

	return modified, nil
}
