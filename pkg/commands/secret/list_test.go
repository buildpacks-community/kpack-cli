// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	secretcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/secret"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestSecretListCommand(t *testing.T) {
	spec.Run(t, "TestSecretListCommand", testSecretListCommand)
}

func testSecretListCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
	)

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeK8sProvider(k8sClient, defaultNamespace)
		return secretcmds.NewListCommand(clientSetProvider)
	}

	when("listing secrets", func() {
		when("listing secrets in the default namespace", func() {
			when("there are secrets in the default service account", func() {
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

					const expectedOutput = `NAME            TARGET                         AVAILABLE
secret-one      https://index.docker.io/v1/    false
secret-three                                   false
secret-two      some-git-url                   false

`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							serviceAccount,
						},
						ExpectedOutput: expectedOutput,
					}.TestK8s(t, cmdFunc)
				})
			})

			when("there are secrets in a custom service account", func() {
				it("lists the secrets", func() {
					serviceAccount := &corev1.ServiceAccount{
						ObjectMeta: v1.ObjectMeta{
							Name:      "some-sa",
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

					const expectedOutput = `NAME            TARGET                         AVAILABLE
secret-one      https://index.docker.io/v1/    false
secret-three                                   false
secret-two      some-git-url                   false

`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							serviceAccount,
						},
						Args:           []string{"--service-account", "some-sa"},
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
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no secrets found in \"some-default-namespace\" namespace for \"default\" service account\n",
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

					const expectedOutput = `NAME            TARGET                         AVAILABLE
secret-one      https://index.docker.io/v1/    false
secret-three                                   false
secret-two      some-git-url                   false

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
						Args:                []string{"-n", namespace},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no secrets found in \"some-namespace\" namespace for \"default\" service account\n",
					}.TestK8s(t, cmdFunc)
				})
			})
		})
	})
}
