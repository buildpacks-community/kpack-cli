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
		Use:   "create <name>",
		Short: "Create a secret configuration",
		Long: `Create a secret configuration using registry or git credentials.

The flags for this command determine the type of secret that will be created:

	"--dockerhub" to create DockerHub credentials

	"--gcr" to create Google Container Registry credentials

	"--registry" and "--registry-user" to create credentials for other registries

	"--git" and "--git-ssh-key" to create SSH based git credentials

	"--git" and "--git-user" to create Basic Auth based git credentials`,
		Example: `tbctl secret create my-docker-hub-creds --dockerhub dockerhub-id
tbctl secret create my-gcr-creds --gcr /path/to/gcr/service-account.json
tbctl secret create my-registry-cred --registry example-registry.io/my-repo --registry-user my-registry-user
tbctl secret create my-git-ssh-cred --git git@github.com --git-ssh-key /path/to/git/ssh-private-key.pem
tbctl secret create my-git-cred --git https://github.com --git-user my-git-user`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			sec, err := secretFactory.MakeSecret(args[0], namespace)
			if err != nil {
				return err
			}

			_, err = k8sClient.CoreV1().Secrets(namespace).Create(sec)
			if err != nil {
				return err
			}

			serviceAccount, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get("default", metav1.GetOptions{})
			if err != nil {
				return err
			}
			serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{Name: args[0]})
			serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: args[0]})

			_, err = k8sClient.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")
	cmd.Flags().StringVarP(&secretFactory.DockerhubId, "dockerhub", "", "", "dockerhub id")
	cmd.Flags().StringVarP(&secretFactory.Registry, "registry", "", "", "registry")
	cmd.Flags().StringVarP(&secretFactory.RegistryUser, "registry-user", "", "", "registry user")
	cmd.Flags().StringVarP(&secretFactory.GcrServiceAccountFile, "gcr", "", "", "path to a file containing the GCR service account")
	cmd.Flags().StringVarP(&secretFactory.Git, "git", "", "", "git url")
	cmd.Flags().StringVarP(&secretFactory.GitSshKeyFile, "git-ssh-key", "", "", "path to a file containing the Git SSH private key")
	cmd.Flags().StringVarP(&secretFactory.GitUser, "git-user", "", "", "git user")

	return cmd
}
