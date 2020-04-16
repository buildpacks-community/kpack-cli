package secret_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/secret"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestSecretDeleteCommand(t *testing.T) {
	spec.Run(t, "TestSecretDeleteCommand", testSecretDeleteCommand)
}

func testSecretDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
	)

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		return secretcmds.NewDeleteCommand(k8sClient, defaultNamespace)
	}

	when("deleting secrets", func() {
		when("deleting secrets in the default namespace", func() {
			const secretName = "some-secret"

			when("the secret exist", func() {
				it("deletes the secrets", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      secretName,
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								secret.RegistryAnnotation: secret.DockerhubUrl,
							},
						},
					}

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
						},
						Args:           []string{secretName},
						ExpectedOutput: "\"some-secret\" deleted\n",
						ExpectDeletes: []clientgotesting.DeleteActionImpl{
							{
								ActionImpl: clientgotesting.ActionImpl{
									Namespace: defaultNamespace,
								},
								Name: secretName,
							},
						},
					}.TestK8s(t, cmdFunc)
				})
			})

			when("the secret does not exist", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args: []string{"some-secret"},
						ExpectDeletes: []clientgotesting.DeleteActionImpl{
							{
								ActionImpl: clientgotesting.ActionImpl{
									Namespace: defaultNamespace,
								},
								Name: secretName,
							},
						},
						ExpectedOutput: "Error: secrets \"some-secret\" not found\n",
						ExpectErr:      true,
					}.TestK8s(t, cmdFunc)
				})
			})
		})

		when("deleting secrets in the given namespace", func() {
			const (
				namespace  = "some-namespace"
				secretName = "some-secret"
			)

			when("the secret exist", func() {
				it("deletes the secrets", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
							Annotations: map[string]string{
								secret.RegistryAnnotation: secret.DockerhubUrl,
							},
						},
					}

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
						},
						Args:           []string{secretName, "-n", namespace},
						ExpectedOutput: "\"some-secret\" deleted\n",
						ExpectDeletes: []clientgotesting.DeleteActionImpl{
							{
								ActionImpl: clientgotesting.ActionImpl{
									Namespace: namespace,
								},
								Name: secretName,
							},
						},
					}.TestK8s(t, cmdFunc)
				})
			})

			when("the secret does not exist", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args: []string{secretName, "-n", namespace},
						ExpectDeletes: []clientgotesting.DeleteActionImpl{
							{
								ActionImpl: clientgotesting.ActionImpl{
									Namespace: namespace,
								},
								Name: secretName,
							},
						},
						ExpectedOutput: "Error: secrets \"some-secret\" not found\n",
						ExpectErr:      true,
					}.TestK8s(t, cmdFunc)
				})
			})
		})
	})
}
