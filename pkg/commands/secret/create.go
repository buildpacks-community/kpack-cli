// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret

import (
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/secret"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, secretFactory *secret.Factory) *cobra.Command {
	var (
		namespace string
		dryRun    bool
		output    string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a secret configuration",
		Long: `Create a secret configuration using registry or git credentials in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine the type of secret that will be created:

  "--dockerhub" to create DockerHub credentials.
  Use the "DOCKER_PASSWORD" env var to bypass the password prompt.

  "--gcr" to create Google Container Registry credentials.
  Alternatively, provided the credentials in the "GCR_SERVICE_ACCOUNT_PATH" env var instead of the "--gcr" flag.

  "--registry" and "--registry-user" to create credentials for other registries.
  Use the "REGISTRY_PASSWORD" env var to bypass the password prompt.

  "--git-url" and "--git-ssh-key" to create SSH based git credentials.
  "--git-url" should not contain the repository path (eg. git@github.com not git@github.com:my/repo)
  Alternatively, provided the credentials in the "GIT_SSH_KEY_PATH" env var instead of the "--git-ssh-key" flag.

  "--git-url" and "--git-user" to create Basic Auth based git credentials.
  "--git-url" should not contain the repository path (eg. https://github.com not https://github.com/my/repo) 
  Use the "GIT_PASSWORD" env var to bypass the password prompt.`,
		Example: `kp secret create my-docker-hub-creds --dockerhub dockerhub-id
kp secret create my-gcr-creds --gcr /path/to/gcr/service-account.json
kp secret create my-registry-cred --registry example-registry.io/my-repo --registry-user my-registry-user
kp secret create my-git-ssh-cred --git-url git@github.com --git-ssh-key /path/to/git/ssh-private-key.pem
kp secret create my-git-cred --git-url https://github.com --git-user my-git-user`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			if val, ok := os.LookupEnv("GCR_SERVICE_ACCOUNT_PATH"); ok {
				secretFactory.GcrServiceAccountFile = val
			}

			if val, ok := os.LookupEnv("GIT_SSH_KEY_PATH"); ok {
				secretFactory.GitSshKeyFile = val
			}

			secret, target, err := secretFactory.MakeSecret(args[0], cs.Namespace)
			if err != nil {
				return err
			}

			if !ch.IsDryRun() {
				secret, err = cs.K8sClient.CoreV1().Secrets(cs.Namespace).Create(secret)
				if err != nil {
					return err
				}
			}

			err = ch.PrintObj(secret)
			if err != nil {
				return err
			}

			serviceAccount, err := cs.K8sClient.CoreV1().ServiceAccounts(cs.Namespace).Get("default", metav1.GetOptions{})
			if err != nil {
				return err
			}

			serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{Name: args[0]})

			if secret.Type == corev1.SecretTypeDockerConfigJson {
				serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: args[0]})
			}

			err = updateManagedSecretsAnnotation(err, serviceAccount, args[0], target)
			if err != nil {
				return err
			}

			if !ch.IsDryRun() {
				serviceAccount, err = cs.K8sClient.CoreV1().ServiceAccounts(cs.Namespace).Update(serviceAccount)
				if err != nil {
					return err
				}
			}

			err = ch.PrintObj(secret)
			if err != nil {
				return err
			}

			return ch.PrintResult("%q created", secret.Name)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&secretFactory.DockerhubId, "dockerhub", "", "", "dockerhub id")
	cmd.Flags().StringVarP(&secretFactory.Registry, "registry", "", "", "registry")
	cmd.Flags().StringVarP(&secretFactory.RegistryUser, "registry-user", "", "", "registry user")
	cmd.Flags().StringVarP(&secretFactory.GcrServiceAccountFile, "gcr", "", "", "path to a file containing the GCR service account")
	cmd.Flags().StringVarP(&secretFactory.GitUrl, "git-url", "", "", "git url")
	cmd.Flags().StringVarP(&secretFactory.GitSshKeyFile, "git-ssh-key", "", "", "path to a file containing the GitUrl SSH private key")
	cmd.Flags().StringVarP(&secretFactory.GitUser, "git-user", "", "", "git user")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&output, "output", "", "output format. supported formats are: yaml, json")
	return cmd
}

func updateManagedSecretsAnnotation(err error, sa *corev1.ServiceAccount, name, target string) error {
	managedSecrets, err := readManagedSecrets(sa)
	if err != nil {
		return err
	}

	managedSecrets[name] = target

	return writeManagedSecrets(managedSecrets, sa)
}
