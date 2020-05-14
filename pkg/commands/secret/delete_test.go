package secret_test

import (
	"fmt"
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
		contextProvider := testhelpers.NewFakeK8sContextProvider(defaultNamespace, k8sClient)
		return secretcmds.NewDeleteCommand(contextProvider)
	}

	when("deleting secrets", func() {
		when("deleting secrets in the default namespace", func() {
			const secretName = "some-secret"

			when("the secret exist", func() {
				it("deletes the secret and removes it from the default service account", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      secretName,
							Namespace: defaultNamespace,
						},
					}

					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s", "foo":"bar"}`, secretName, secret.DockerhubUrl),
							},
						},
						Secrets: []corev1.ObjectReference{
							{Name: secretName},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: secretName},
						},
					}

					expectedServiceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								secretcmds.ManagedSecretAnnotationKey: `{"foo":"bar"}`,
							},
						},
					}

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							serviceAccount,
						},
						Args:           []string{secretName},
						ExpectedOutput: "\"some-secret\" deleted\n",
						ExpectUpdates: []clientgotesting.UpdateActionImpl{
							{
								Object: expectedServiceAccount,
							},
						},
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
				it("deletes the secret and removes it from the default service account", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
						},
					}

					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: namespace,
							Annotations: map[string]string{
								secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s", "foo":"bar"}`, secretName, secret.DockerhubUrl),
							},
						},
						Secrets: []corev1.ObjectReference{
							{Name: secretName},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: secretName},
						},
					}

					expectedServiceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "default",
							Namespace: namespace,
							Annotations: map[string]string{
								secretcmds.ManagedSecretAnnotationKey: `{"foo":"bar"}`,
							},
						},
					}

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							serviceAccount,
						},
						Args:           []string{secretName, "-n", namespace},
						ExpectedOutput: "\"some-secret\" deleted\n",
						ExpectUpdates: []clientgotesting.UpdateActionImpl{
							{
								Object: expectedServiceAccount,
							},
						},
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
