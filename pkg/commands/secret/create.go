package secret

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	DockerhubUrl       = "https://index.docker.io/v1/"
	RegistryAnnotation = "build.pivotal.io/docker"
)

type PasswordReader interface {
	Read(out io.Writer, prompt, envVar string) (string, error)
}

func NewCreateCommand(k8sClient k8s.Interface, passwordReader PasswordReader, defaultNamespace string) *cobra.Command {
	var (
		dockerhubId string
		namespace   string
	)

	cmd := &cobra.Command{
		Use:          "create <name>",
		Short:        "Create a secret configuration",
		Long:         "Create a secret configuration using registry or github credentials.",
		Example:      "tbctl secret create my-docker-hub-creds --dockerhub dockerhub-id",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			password, err := passwordReader.Read(cmd.OutOrStdout(), "dockerhub password: ", "DOCKER_PASSWORD")
			if err != nil {
				return err
			}

			configJson := dockerConfigJson{Auths: DockerCreds{
				DockerhubUrl: authn.AuthConfig{
					Username: dockerhubId,
					Password: password,
				},
			}}
			dockerCfgJson, err := json.Marshal(configJson)
			if err != nil {
				return err
			}

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      args[0],
					Namespace: namespace,
					Annotations: map[string]string{
						RegistryAnnotation: DockerhubUrl,
					},
				},
				Data: map[string][]byte{
					corev1.DockerConfigJsonKey: dockerCfgJson,
				},
				Type: corev1.SecretTypeDockerConfigJson,
			}

			_, err = k8sClient.CoreV1().Secrets(namespace).Create(secret)
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

			_, err = cmd.OutOrStdout().Write([]byte(fmt.Sprintf("\"%s\" created\n", args[0])))
			return err
		},
	}

	cmd.Flags().StringVarP(&dockerhubId, "dockerhub", "", "", "dockerhub id")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "the namespace of the image")

	return cmd
}
