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
					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								secretcmds.ManagedSecretAnnotationKey: `{"secret-one":"https://index.docker.io/v1/", "secret-two":"some-git-url", "secret-three":""}`,
							},
						},
						Secrets: []corev1.ObjectReference{
							{
								Name: "secret-one",
							},
							{
								Name: "secret-two",
							},
							{
								Name: "secret-three",
							},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{
								Name: "secret-one",
							},
						},
					}

					const expectedOutput = `NAME            TARGET
secret-one      https://index.docker.io/v1/
secret-three    
secret-two      some-git-url
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							serviceAccount,
						},
						ExpectedOutput: expectedOutput,
					}.TestK8s(t, cmdFunc)
				})
			})

			when("there are no secrets", func() {
				it("prints an appropriate message", func() {
					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: defaultNamespace,
						},
					}

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							serviceAccount,
						},
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
					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: namespace,
							Annotations: map[string]string{
								secretcmds.ManagedSecretAnnotationKey: `{"secret-one":"https://index.docker.io/v1/", "secret-two":"some-git-url", "secret-three":""}`,
							},
						},
						Secrets: []corev1.ObjectReference{
							{
								Name: "secret-one",
							},
							{
								Name: "secret-two",
							},
							{
								Name: "secret-three",
							},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{
								Name: "secret-one",
							},
						},
					}

					const expectedOutput = `NAME            TARGET
secret-one      https://index.docker.io/v1/
secret-three    
secret-two      some-git-url
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							serviceAccount,
						},
						Args:           []string{"-n", namespace},
						ExpectedOutput: expectedOutput,
					}.TestK8s(t, cmdFunc)
				})
			})

			when("there are no secrets", func() {
				it("prints an appropriate message", func() {
					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: namespace,
						},
					}

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							serviceAccount,
						},
						Args:           []string{"-n", namespace},
						ExpectErr:      true,
						ExpectedOutput: "Error: no secrets found in \"some-namespace\" namespace\n",
					}.TestK8s(t, cmdFunc)
				})
			})
		})
	})
}
