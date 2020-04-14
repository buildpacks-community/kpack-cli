package secret

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"
	"gopkg.in/errgo.v2/fmt/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	DockerhubUrl       = "https://index.docker.io/v1/"
	GcrUrl             = "gcr.io"
	GcrUser            = "_json_key"
	RegistryAnnotation = "build.pivotal.io/docker"
	GitAnnotation      = "build.pivotal.io/git"
)

type PasswordReader interface {
	Read(out io.Writer, prompt, envVar string) (string, error)
}

func NewCreateCommand(k8sClient k8s.Interface, passwordReader PasswordReader, defaultNamespace string) *cobra.Command {
	var (
		credFactory credentialFactory
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
			cred, err := credFactory.makeCredential(cmd.OutOrStdout(), passwordReader)
			if err != nil {
				return err
			}

			secret, err := cred.toSecret(args[0], namespace)
			if err != nil {
				return err
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

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s created\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVarP(&credFactory.dockerhubId, "dockerhub", "", "", "dockerhub id")
	cmd.Flags().StringVarP(&credFactory.registry, "registry", "", "", "registry")
	cmd.Flags().StringVarP(&credFactory.registryUser, "registry-user", "", "", "registry user")
	cmd.Flags().StringVarP(&credFactory.gcrServiceAccountFile, "gcr", "", "", "path to a file containing the GCR service account")
	cmd.Flags().StringVarP(&credFactory.git, "git", "", "", "git url")
	cmd.Flags().StringVarP(&credFactory.gitSshKeyFile, "git-ssh-key", "", "", "path to a file containing the Git SSH private key")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "the namespace of the image")

	return cmd
}

type secretKind int

const (
	registryKind secretKind = iota
	gitSshKind
)

type credential struct {
	kind     secretKind
	resource string
	username string
	password string
}

func (c credential) toSecret(name, namespace string) (*corev1.Secret, error) {
	if c.kind == registryKind {
		configJson := dockerConfigJson{Auths: DockerCreds{
			c.resource: authn.AuthConfig{
				Username: c.username,
				Password: c.password,
			},
		}}
		dockerCfgJson, err := json.Marshal(configJson)
		if err != nil {
			return nil, err
		}

		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Annotations: map[string]string{
					RegistryAnnotation: c.resource,
				},
			},
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: dockerCfgJson,
			},
			Type: corev1.SecretTypeDockerConfigJson,
		}, nil
	} else {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Annotations: map[string]string{
					GitAnnotation: c.resource,
				},
			},
			Data: map[string][]byte{
				corev1.SSHAuthPrivateKey: []byte(c.password),
			},
			Type: corev1.SecretTypeSSHAuth,
		}, nil
	}
}

type credentialFactory struct {
	dockerhubId           string
	registry              string
	registryUser          string
	gcrServiceAccountFile string
	git                   string
	gitSshKeyFile         string
}

func (c credentialFactory) makeCredential(writer io.Writer, passwordReader PasswordReader) (credential, error) {
	if c.dockerhubId != "" {
		password, err := passwordReader.Read(writer, "dockerhub password: ", "DOCKER_PASSWORD")
		if err != nil {
			return credential{}, err
		}

		return credential{
			kind:     registryKind,
			resource: DockerhubUrl,
			username: c.dockerhubId,
			password: password,
		}, nil
	} else if c.registry != "" && c.registryUser != "" {
		password, err := passwordReader.Read(writer, "registry password: ", "REGISTRY_PASSWORD")
		if err != nil {
			return credential{}, err
		}

		return credential{
			kind:     registryKind,
			resource: c.registry,
			username: c.registryUser,
			password: password,
		}, nil
	} else if c.gcrServiceAccountFile != "" {
		buf, err := ioutil.ReadFile(c.gcrServiceAccountFile)
		if err != nil {
			return credential{}, err
		}

		return credential{
			kind:     registryKind,
			resource: GcrUrl,
			username: GcrUser,
			password: string(buf),
		}, nil
	} else if c.git != "" && c.gitSshKeyFile != "" {
		buf, err := ioutil.ReadFile(c.gitSshKeyFile)
		if err != nil {
			return credential{}, err
		}

		return credential{
			kind:     gitSshKind,
			resource: c.git,
			username: "",
			password: string(buf),
		}, nil
	}

	return credential{}, errors.Newf("incorrect flags provided")
}
