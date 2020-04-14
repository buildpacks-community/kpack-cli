package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/pivotal/build-service-cli/pkg/secret"
)

func NewCreateCommand(k8sClient k8s.Interface, secretFactory *secret.Factory, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:          "create <name>",
		Short:        "Create a secret configuration",
		Long:         "Create a secret configuration using registry or github credentials.",
		Example:      "tbctl secret create my-docker-hub-creds --dockerhub dockerhub-id",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cred, err := secretFactory.MakeSecret(args[0], namespace)
			if err != nil {
				return err
			}

			_, err = k8sClient.CoreV1().Secrets(namespace).Create(cred)
			if err != nil {
				return err
			}

			serviceAccount, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get("default", metav1.GetOptions{})
			if err != nil {
				return err
			}

			serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: args[0]})
			serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{Name: args[0]})

			_, err = k8sClient.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s created\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVarP(&secretFactory.DockerhubId, "dockerhub", "", "", "dockerhub id")
	cmd.Flags().StringVarP(&secretFactory.Registry, "registry", "", "", "registry")
	cmd.Flags().StringVarP(&secretFactory.RegistryUser, "registry-user", "", "", "registry user")
	cmd.Flags().StringVarP(&secretFactory.GcrServiceAccountFile, "gcr", "", "", "path to a file containing the GCR service account")
	cmd.Flags().StringVarP(&secretFactory.Git, "git", "", "", "git url")
	cmd.Flags().StringVarP(&secretFactory.GitSshKeyFile, "git-ssh-key", "", "", "path to a file containing the Git SSH private key")
	cmd.Flags().StringVarP(&secretFactory.GitUser, "git-user", "", "", "git user")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "the namespace of the image")

	return cmd
}
