package secret_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"gopkg.in/errgo.v2/fmt/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestSecretCreateCommand(t *testing.T) {
	spec.Run(t, "TestSecretCreateCommand", testSecretCreateCommand)
}

func testSecretCreateCommand(t *testing.T, when spec.G, it spec.S) {

	const (
		defaultNamespace = "some-default-namespace"
	)

	passwordReader := fakePasswordReader{
		passwords: map[string]string{},
	}

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		return secret.NewCreateCommand(k8sClient, passwordReader, defaultNamespace)
	}

	when("creating a dockerhub secret", func() {

		var (
			dockerhubId          = "my-dockerhub-id"
			dockerPassword       = "dummy-password"
			secretName           = "my-docker-cred"
			expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
		)

		passwordReader.passwords["DOCKER_PASSWORD"] = dockerPassword

		when("a namespace is not provided", func() {
			it("creates a secret with the correct annotations for docker in the default namespace and updates the service account", func() {

				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secret.RegistryAnnotation: secret.DockerhubUrl,
						},
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedDockerConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--dockerhub", dockerhubId},
					ExpectedOutput: `my-docker-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("a namespace is provided", func() {

			var (
				namespace = "some-namespace"
			)

			it("creates a secret with the correct annotations for docker in the provided namespace and updates the service account", func() {

				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
						Annotations: map[string]string{
							secret.RegistryAnnotation: secret.DockerhubUrl,
						},
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedDockerConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--dockerhub", dockerhubId, "-n", namespace},
					ExpectedOutput: `my-docker-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("creating a generic registry secret", func() {
		var (
			registry               = "my-registry.io/my-repo"
			registryUser           = "my-registry-user"
			registryPassword       = "dummy-password"
			secretName             = "my-registry-cred"
			expectedRegistryConfig = fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", registry, registryUser, registryPassword)
		)

		passwordReader.passwords["REGISTRY_PASSWORD"] = registryPassword

		when("a namespace is not provided", func() {
			it("creates a secret with the correct annotations for the registry in the default namespace and updates the service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secret.RegistryAnnotation: registry,
						},
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--registry", registry, "--registry-user", registryUser},
					ExpectedOutput: `my-registry-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("a namespace is provided", func() {

			var (
				namespace = "some-namespace"
			)

			it("creates a secret with the correct annotations for the registry in the provided namespace and updates the service account", func() {

				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
						Annotations: map[string]string{
							secret.RegistryAnnotation: registry,
						},
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--registry", registry, "--registry-user", registryUser, "-n", namespace},
					ExpectedOutput: `my-registry-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("creating a gcr registry secret", func() {
		var (
			gcrServiceAccountFile  = "./testdata/gcr-service-account.json"
			secretName             = "my-gcr-cred"
			expectedRegistryConfig = fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"{\"some-key\":\"some-value\"}"}}}`, secret.GcrUrl, secret.GcrUser)
		)

		when("a namespace is not provided", func() {
			it("creates a secret with the correct annotations for gcr in the default namespace and updates the service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secret.RegistryAnnotation: secret.GcrUrl,
						},
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--gcr", gcrServiceAccountFile},
					ExpectedOutput: `my-gcr-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("a namespace is provided", func() {

			var (
				namespace = "some-namespace"
			)

			it("creates a secret with the correct annotations for the registry in the provided namespace and updates the service account", func() {

				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
						Annotations: map[string]string{
							secret.RegistryAnnotation: secret.GcrUrl,
						},
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--gcr", gcrServiceAccountFile, "-n", namespace},
					ExpectedOutput: `my-gcr-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("creating a git ssh secret", func() {
		var (
			gitRepo    = "git@github.com"
			gitSshFile = "./testdata/git-ssh.pem"
			secretName = "my-git-ssh-cred"
		)

		when("a namespace is not provided", func() {
			it("creates a secret with the correct annotations for git ssh in the default namespace and updates the service account", func() {
				expectedGitSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secret.GitAnnotation: gitRepo,
						},
					},
					Data: map[string][]byte{
						corev1.SSHAuthPrivateKey: []byte("some git ssh data"),
					},
					Type: corev1.SecretTypeSSHAuth,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--git", gitRepo, "--git-ssh-key", gitSshFile},
					ExpectedOutput: `my-git-ssh-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("a namespace is provided", func() {
			var (
				namespace = "some-namespace"
			)

			it("creates a secret with the correct annotations for git ssh in the provided namespace and updates the service account", func() {
				expectedGitSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
						Annotations: map[string]string{
							secret.GitAnnotation: gitRepo,
						},
					},
					Data: map[string][]byte{
						corev1.SSHAuthPrivateKey: []byte("some git ssh data"),
					},
					Type: corev1.SecretTypeSSHAuth,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
				}

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--git", gitRepo, "--git-ssh-key", gitSshFile, "-n", namespace},
					ExpectedOutput: `my-git-ssh-cred created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})
}

type fakePasswordReader struct {
	passwords map[string]string
}

func (s fakePasswordReader) Read(_ io.Writer, _, envVar string) (string, error) {
	if password, ok := s.passwords[envVar]; ok {
		return password, nil
	}
	return "", errors.Newf("secret for %s not found", envVar)
}
