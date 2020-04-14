package secret_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
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

	passwordReader := stubPasswordReader{}

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		return secret.NewCreateCommand(k8sClient, passwordReader, defaultNamespace)
	}

	when("creating a dockerhub secret", func() {

		it("creates a secret with the correct annotations for docker in the default namespace and updates the service account", func() {

			var (
				dockerhubId          = "my-dockerhub-id"
				dockerPassword       = "dummy-password"
				secretName           = "my-docker-cred"
				expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
			)

			passwordReader.password = dockerPassword

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
}

type stubPasswordReader struct {
	password string
}

func (s stubPasswordReader) Read(_ io.Writer, _, _ string) (string, error) {
	return s.password, nil
}
