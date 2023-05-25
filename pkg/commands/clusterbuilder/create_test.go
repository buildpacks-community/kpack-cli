// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	cbcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuilder"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuilderCreateCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderCreateCommand", testCreateCommand(cbcmds.NewCreateCommand))
}

func testCreateCommand(clusterBuilderCommand func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		var (
			config = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository":                          "default-registry.io/default-repo",
					"default.repository.serviceaccount":           "some-serviceaccount",
					"default.repository.serviceaccount.namespace": "some-namespace",
				},
			}

			expectedBuilder = &v1alpha2.ClusterBuilder{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha2.ClusterBuilderKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-builder",
					Annotations: map[string]string{
						"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{}}}`,
					},
				},
				Spec: v1alpha2.ClusterBuilderSpec{
					BuilderSpec: v1alpha2.BuilderSpec{
						Tag: "default-registry.io/default-repo:clusterbuilder-test-builder",
						Stack: corev1.ObjectReference{
							Name: "some-stack",
							Kind: v1alpha2.ClusterStackKind,
						},
						Store: corev1.ObjectReference{
							Name: "some-store",
							Kind: v1alpha2.ClusterStoreKind,
						},
						Order: []v1alpha2.BuilderOrderEntry{
							{
								Group: []v1alpha2.BuilderBuildpackRef{
									{
										BuildpackRef: corev1alpha1.BuildpackRef{
											BuildpackInfo: corev1alpha1.BuildpackInfo{
												Id: "org.cloudfoundry.nodejs",
											},
										},
									},
								},
							},
							{
								Group: []v1alpha2.BuilderBuildpackRef{
									{
										BuildpackRef: corev1alpha1.BuildpackRef{
											BuildpackInfo: corev1alpha1.BuildpackInfo{
												Id: "org.cloudfoundry.go",
											},
										},
									},
								},
							},
						},
					},
					ServiceAccountRef: corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}
		)

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
			return clusterBuilderCommand(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("creates a ClusterBuilder", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("creates a ClusterBuilder with the default stack", func() {
			expectedBuilder.Spec.Stack.Name = "default"
			expectedBuilder.Spec.Store.Name = "default"
			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-test-builder","stack":{"kind":"ClusterStack","name":"default"},"store":{"kind":"ClusterStore","name":"default"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("creates a ClusterBuilder with the default tag when tag is not specified", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					expectedBuilder.Name,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("fails when tag is not specified and default.repository key is not found in kp-config configmap", func() {
			badConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository.serviceaccount": "some-serviceaccount",
				},
			}

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					badConfig,
				},
				Args: []string{
					expectedBuilder.Name,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: default-registry.io/default-repo:clusterbuilder-test-builder
status:
  stack: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectCreates: []runtime.Object{
						expectedBuilder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-builder",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"test-builder\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-test-builder\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"org.cloudfoundry.nodejs\"}]},{\"group\":[{\"id\":\"org.cloudfoundry.go\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "default-registry.io/default-repo:clusterbuilder-test-builder",
        "stack": {
            "kind": "ClusterStack",
            "name": "some-stack"
        },
        "store": {
            "kind": "ClusterStore",
            "name": "some-store"
        },
        "order": [
            {
                "group": [
                    {
                        "id": "org.cloudfoundry.nodejs"
                    }
                ]
            },
            {
                "group": [
                    {
                        "id": "org.cloudfoundry.go"
                    }
                ]
            }
        ],
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {}
    }
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectCreates: []runtime.Object{
						expectedBuilder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a ClusterBuilder and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"--dry-run",
					},
					ExpectedOutput: `ClusterBuilder "test-builder" created (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: default-registry.io/default-repo:clusterbuilder-test-builder
status:
  stack: {}
`

				it("does not create a ClusterBuilder and prints the resource output", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
						},
						Args: []string{
							expectedBuilder.Name,
							"--tag", expectedBuilder.Spec.Tag,
							"--stack", expectedBuilder.Spec.Stack.Name,
							"--store", expectedBuilder.Spec.Store.Name,
							"--order", "./testdata/order.yaml",
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("buildpack flag is used", func() {
			it("creates a builder using the buildpack flag", func() {

				expectedBuilder.Spec.Order = []v1alpha2.BuilderOrderEntry{
					{
						Group: []v1alpha2.BuilderBuildpackRef{
							{
								BuildpackRef: corev1alpha1.BuildpackRef{
									BuildpackInfo: corev1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
							{
								BuildpackRef: corev1alpha1.BuildpackRef{
									BuildpackInfo: corev1alpha1.BuildpackInfo{
										Id:      "org.cloudfoundry.nodejs",
										Version: "1",
									},
								},
							},
							{
								BuildpackRef: corev1alpha1.BuildpackRef{
									BuildpackInfo: corev1alpha1.BuildpackInfo{
										Id:      "org.cloudfoundry.ruby",
										Version: "1.2.3",
									},
								},
							},
						},
					},
				}
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.go"},{"id":"org.cloudfoundry.nodejs","version":"1"},{"id":"org.cloudfoundry.ruby","version":"1.2.3"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--buildpack", "org.cloudfoundry.go,org.cloudfoundry.nodejs@1",
						"--buildpack", "org.cloudfoundry.ruby@1.2.3",
					},
					ExpectedOutput: `ClusterBuilder "test-builder" created
`,
					ExpectCreates: []runtime.Object{
						expectedBuilder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("buildpack and order flags are used together", func() {
				it("returns an error", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
						},
						Args: []string{
							expectedBuilder.Name,
							"--tag", expectedBuilder.Spec.Tag,
							"--order", "./testdata/order.yaml",
							"--buildpack", "some-buildpack-name",
						},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: cannot use --order and --buildpack together\n",
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	}
}
