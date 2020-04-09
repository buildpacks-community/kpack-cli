package secret_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestSecretCreateCommand(t *testing.T) {
	spec.Run(t, "TestSecretCreateCommand", testSecretCreateCommand)
}

func testSecretCreateCommand(t *testing.T, when spec.G, it spec.S) {

	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return secret.NewCreateCommand(clientSet, defaultNamespace)
	}

	when("creating a dockerhub secret", func() {

		when("env var DOCKER_PASSWORD is not set", func() {

			it("prompts the user for a password", func() {
			})

			it("creates a secret with the correct annotations for docker", func() {
			})
		})

		when("env var DOCKER_PASSWORD is set", func() {
			it("creates a secret with the correct annotations for docker in the default namespace and updates the service account", func() {
				err := os.Setenv("DOCKER_PASSWORD", "dummy-password")
				require.NoError(t, err)

				dockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-docker-cred",
						Namespace: defaultNamespace,
					},
					StringData: map[string]string{
						corev1.DockerConfigJsonKey: "foo",
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args:           []string{"--dockerhub", "my-dockerhub-id", "my-docker-cred"},
					ExpectedOutput: "\"my-docker-cred\" created",
					ExpectCreates: []runtime.Object{
						dockerSecret,
					},
				}.TestK8s(t, cmdFunc)

			})
		})
	})

}
