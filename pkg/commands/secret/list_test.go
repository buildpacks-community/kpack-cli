package secret_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/secret"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestSecretListCommand(t *testing.T) {
	spec.Run(t, "TestSecretListCommand", testSecretListCommand)
}

func testSecretListCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
	)

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		return secretcmds.NewListCommand(k8sClient, defaultNamespace)
	}

	when("listing secrets", func() {
		when("listing secrets in the default namespace", func() {
			when("there are secrets", func() {
				it("lists the secrets", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-one",
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								secret.TargetAnnotation: secret.DockerhubUrl,
							},
						},
					}
					secretTwo := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-two",
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								secret.GitAnnotation: "some-git-url",
							},
						},
					}
					secretThree := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-three",
							Namespace: defaultNamespace,
						},
					}
					secretFour := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-four",
							Namespace: "other-namespace",
						},
					}

					const expectedOutput = `NAME            TARGET
secret-one      https://index.docker.io/v1/
secret-two      some-git-url
secret-three    
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							secretTwo,
							secretThree,
							secretFour,
						},
						ExpectedOutput: expectedOutput,
					}.TestK8s(t, cmdFunc)
				})
			})

			when("there are no secrets", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						ExpectErr:      true,
						ExpectedOutput: "Error: no secrets found in \"some-default-namespace\" namespace\n",
					}.TestK8s(t, cmdFunc)
				})
			})
		})

		when("listing secrets in a given namespace", func() {
			const namespace = "some-namespace"

			when("there are secrets", func() {
				it("lists the secrets", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-one",
							Namespace: namespace,
							Annotations: map[string]string{
								secret.TargetAnnotation: secret.DockerhubUrl,
							},
						},
					}
					secretTwo := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-two",
							Namespace: namespace,
							Annotations: map[string]string{
								secret.GitAnnotation: "some-git-url",
							},
						},
					}
					secretThree := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-three",
							Namespace: namespace,
						},
					}
					secretFour := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "secret-four",
							Namespace: defaultNamespace,
						},
					}

					const expectedOutput = `NAME            TARGET
secret-one      https://index.docker.io/v1/
secret-two      some-git-url
secret-three    
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							secretTwo,
							secretThree,
							secretFour,
						},
						Args:           []string{"-n", namespace},
						ExpectedOutput: expectedOutput,
					}.TestK8s(t, cmdFunc)
				})
			})

			when("there are no secrets", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{"-n", namespace},
						ExpectErr:      true,
						ExpectedOutput: "Error: no secrets found in \"some-namespace\" namespace\n",
					}.TestK8s(t, cmdFunc)
				})
			})
		})
	})
}
