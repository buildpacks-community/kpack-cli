package secret_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
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

	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		return secret.NewCreateCommand(k8sClient, defaultNamespace)
	}

	when("creating a dockerhub secret", func() {

		when("env var DOCKER_PASSWORD is not set", func() {

			it("prompts the user for a password and creates a secret with the correct annotations for docker", func() {

				var (
					dockerhubId          = "my-dockerhub-id"
					dockerPassword       = "dummy-password"
					secretName           = "my-docker-cred"
					expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
				)

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

				expectedPasswordPrompt := "dockerhub password: "
				expectedCommandOutput := `"my-docker-cred" created
`

				testhelpers.CommandTest{
					StringInput: "dummy-password",
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args:           []string{"--dockerhub", dockerhubId, "my-docker-cred"},
					ExpectedOutput: expectedPasswordPrompt + expectedCommandOutput,
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

		when("env var DOCKER_PASSWORD is set", func() {
			it("creates a secret with the correct annotations for docker in the default namespace and updates the service account", func() {

				var (
					dockerhubId          = "my-dockerhub-id"
					dockerPassword       = "dummy-password"
					secretName           = "my-docker-cred"
					expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
				)

				err := os.Setenv("DOCKER_PASSWORD", dockerPassword)
				require.NoError(t, err)

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
					Args: []string{"--dockerhub", dockerhubId, "my-docker-cred"},
					ExpectedOutput: `"my-docker-cred" created
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
}
