// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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

	secretcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/secret"
	"github.com/vmware-tanzu/kpack-cli/pkg/secret"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestSecretDeleteCommand(t *testing.T) {
	spec.Run(t, "TestSecretDeleteCommand", testSecretDeleteCommand)
}

func testSecretDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
	)

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeK8sProvider(k8sClient, defaultNamespace)
		return secretcmds.NewDeleteCommand(clientSetProvider)
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

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							serviceAccount,
						},
						Args: []string{secretName},
						ExpectedOutput: `Secret "some-secret" deleted
`,
						ExpectPatches: []string{
							`{"imagePullSecrets":null,"metadata":{"annotations":{"kpack.io/managedSecret":"{\"foo\":\"bar\"}"}},"secrets":null}`,
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

				it("deletes the secret and removes it from the a custom service account", func() {
					secretOne := &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      secretName,
							Namespace: defaultNamespace,
						},
					}

					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "some-sa",
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

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							serviceAccount,
						},
						Args: []string{secretName, "--service-account", "some-sa"},
						ExpectedOutput: `Secret "some-secret" deleted
`,
						ExpectPatches: []string{
							`{"imagePullSecrets":null,"metadata":{"annotations":{"kpack.io/managedSecret":"{\"foo\":\"bar\"}"}},"secrets":null}`,
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
						ExpectedErrorOutput: "Error: secrets \"some-secret\" not found\n",
						ExpectErr:           true,
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

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							secretOne,
							serviceAccount,
						},
						Args: []string{secretName, "-n", namespace},
						ExpectedOutput: `Secret "some-secret" deleted
`,
						ExpectPatches: []string{
							`{"imagePullSecrets":null,"metadata":{"annotations":{"kpack.io/managedSecret":"{\"foo\":\"bar\"}"}},"secrets":null}`,
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
						ExpectedErrorOutput: "Error: secrets \"some-secret\" not found\n",
						ExpectErr:           true,
					}.TestK8s(t, cmdFunc)
				})
			})
		})
	})
}
