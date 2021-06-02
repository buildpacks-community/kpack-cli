// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuilder"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuilderSaveCommand(t *testing.T) {
	spec.Run(t, "TestBuilderSaveCommand", testClusterBuilderSaveCommand)
}

func testClusterBuilderSaveCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		config = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"canonical.repository":                "some-registry/some-project",
				"canonical.repository.serviceaccount": "some-serviceaccount",
			},
		}

		builder = &v1alpha1.ClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "test-builder",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`,
				},
			},
			Spec: v1alpha1.ClusterBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry/some-project/test-builder",
					Stack: corev1.ObjectReference{
						Name: "some-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "some-store",
						Kind: v1alpha1.ClusterStoreKind,
					},
					Order: []v1alpha1.OrderEntry{
						{
							Group: []v1alpha1.BuildpackRef{
								{
									BuildpackInfo: v1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha1.BuildpackRef{
								{
									BuildpackInfo: v1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountRef: corev1.ObjectReference{
					Namespace: "kpack",
					Name:      "some-serviceaccount",
				},
			},
		}
	)

	fakeCBWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterbuilder.NewSaveCommand(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeCBWaiter
		})
	}

	when("creating", func() {
		it("creates a ClusterBuilder when it does not exist", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					builder.Name,
					"--tag", builder.Spec.Tag,
					"--stack", builder.Spec.Stack.Name,
					"--store", builder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					builder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeCBWaiter.WaitCalls, 1)
		})

		it("creates a ClusterBuilder with the default stack", func() {
			builder.Spec.Stack.Name = "default"
			builder.Spec.Store.Name = "default"
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"default"},"store":{"kind":"ClusterStore","name":"default"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					builder.Name,
					"--tag", builder.Spec.Tag,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					builder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("creates a ClusterBuilder with the canonical tag when tag is not specified", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					builder.Name,
					"--stack", builder.Spec.Stack.Name,
					"--store", builder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					builder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("creates a ClusterBuilder with buildpack flags", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"},{"id":"org.cloudfoundry.go","version":"1.0.1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			builder.Spec.Order = []v1alpha1.OrderEntry{
				{
					Group: []v1alpha1.BuildpackRef{
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id: "org.cloudfoundry.nodejs",
							},
						},
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id:      "org.cloudfoundry.go",
								Version: "1.0.1",
							},
						},
					},
				},
			}

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					builder.Name,
					"--tag", builder.Spec.Tag,
					"--stack", builder.Spec.Stack.Name,
					"--store", builder.Spec.Store.Name,
					"--buildpack", "org.cloudfoundry.nodejs",
					"--buildpack", "org.cloudfoundry.go@1.0.1",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					builder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("returns error when buildpack and order flags are used together", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					builder.Name,
					"--order", "./testdata/patched-order.yaml",
					"--buildpack", "org.cloudfoundry.test-bp",
				},
				ExpectErr:      true,
				ExpectedOutput: "Error: cannot use --order and --buildpack together\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("fails when kp-config map is not found", func() {
			testhelpers.CommandTest{
				Args: []string{
					builder.Name,
					"--tag", builder.Spec.Tag,
					"--stack", builder.Spec.Stack.Name,
					"--store", builder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectErr: true,
				ExpectedOutput: `Error: failed to get canonical service account: configmaps "kp-config" not found
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("fails when canonical.repository.serviceaccount key is not found in kp-config configmap", func() {
			badConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{},
			}

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					badConfig,
				},
				Args: []string{
					builder.Name,
					"--tag", builder.Spec.Tag,
					"--stack", builder.Spec.Stack.Name,
					"--store", builder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectErr: true,
				ExpectedOutput: `Error: failed to get canonical service account: key "canonical.repository.serviceaccount" not found in configmap "kp-config"
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("fails when tag is not specified and canonical.repository key is not found in kp-config configmap", func() {
			badConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"canonical.repository.serviceaccount": "some-serviceaccount",
				},
			}

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					badConfig,
				},
				Args: []string{
					builder.Name,
					"--stack", builder.Spec.Stack.Name,
					"--store", builder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
				},
				ExpectErr: true,
				ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
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
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: some-registry/some-project/test-builder
status:
  stack: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						builder.Name,
						"--tag", builder.Spec.Tag,
						"--stack", builder.Spec.Stack.Name,
						"--store", builder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectCreates: []runtime.Object{
						builder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-builder",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"test-builder\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry/some-project/test-builder\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"org.cloudfoundry.nodejs\"}]},{\"group\":[{\"id\":\"org.cloudfoundry.go\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "some-registry/some-project/test-builder",
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
            "namespace": "kpack",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {}
    }
}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						builder.Name,
						"--tag", builder.Spec.Tag,
						"--stack", builder.Spec.Stack.Name,
						"--store", builder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectCreates: []runtime.Object{
						builder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a ClusterBuilder and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						builder.Name,
						"--tag", builder.Spec.Tag,
						"--stack", builder.Spec.Stack.Name,
						"--store", builder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"--dry-run",
					},
					ExpectedOutput: `ClusterBuilder "test-builder" created (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
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
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: some-registry/some-project/test-builder
status:
  stack: {}
`

				it("does not create a ClusterBuilder and prints the resource output", func() {
					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						Args: []string{
							builder.Name,
							"--tag", builder.Spec.Tag,
							"--stack", builder.Spec.Stack.Name,
							"--store", builder.Spec.Store.Name,
							"--order", "./testdata/order.yaml",
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})

	when("patching", func() {
		it("patches when the ClusterBuilder does exist", func() {
			testhelpers.CommandTest{
				KpackObjects: []runtime.Object{
					builder,
				},
				Args: []string{
					builder.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" patched
`,
				ExpectPatches: []string{
					`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeCBWaiter.WaitCalls, 1)
		})

		it("does not patch if there are no changes", func() {
			testhelpers.CommandTest{
				KpackObjects: []runtime.Object{
					builder,
				},
				Args: []string{
					builder.Name,
				},
				ExpectedOutput: `ClusterBuilder "test-builder" patched (no change)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
spec:
  order:
  - group:
    - id: org.cloudfoundry.test-bp
  - group:
    - id: org.cloudfoundry.fake-bp
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-other-stack
  store:
    kind: ClusterStore
    name: some-other-store
  tag: some-other-tag
status:
  stack: {}
`

				testhelpers.CommandTest{
					KpackObjects: []runtime.Object{
						builder,
					},
					Args: []string{
						builder.Name,
						"--tag", "some-other-tag",
						"--stack", "some-other-stack",
						"--store", "some-other-store",
						"--order", "./testdata/patched-order.yaml",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectPatches: []string{
						`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-builder",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"test-builder\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry/some-project/test-builder\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"org.cloudfoundry.nodejs\"}]},{\"group\":[{\"id\":\"org.cloudfoundry.go\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "some-other-tag",
        "stack": {
            "kind": "ClusterStack",
            "name": "some-other-stack"
        },
        "store": {
            "kind": "ClusterStore",
            "name": "some-other-store"
        },
        "order": [
            {
                "group": [
                    {
                        "id": "org.cloudfoundry.test-bp"
                    }
                ]
            },
            {
                "group": [
                    {
                        "id": "org.cloudfoundry.fake-bp"
                    }
                ]
            }
        ],
        "serviceAccountRef": {
            "namespace": "kpack",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {}
    }
}
`

				testhelpers.CommandTest{
					KpackObjects: []runtime.Object{
						builder,
					},
					Args: []string{
						builder.Name,
						"--tag", "some-other-tag",
						"--stack", "some-other-stack",
						"--store", "some-other-store",
						"--order", "./testdata/patched-order.yaml",
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectPatches: []string{
						`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the patch", func() {
				it("can output unpatched resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
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
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: some-registry/some-project/test-builder
status:
  stack: {}
`

					testhelpers.CommandTest{
						KpackObjects: []runtime.Object{
							builder,
						},
						Args: []string{
							builder.Name,
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
			it("does not patch a ClusterBuilder and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					KpackObjects: []runtime.Object{
						builder,
					},
					Args: []string{
						builder.Name,
						"--tag", "some-other-tag",
						"--stack", "some-other-stack",
						"--store", "some-other-store",
						"--order", "./testdata/patched-order.yaml",
						"--dry-run",
					},
					ExpectedOutput: `ClusterBuilder "test-builder" patched (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the patch", func() {
				it("does not patch and informs of no change", func() {
					testhelpers.CommandTest{
						KpackObjects: []runtime.Object{
							builder,
						},
						Args: []string{
							builder.Name,
							"--dry-run",
						},
						ExpectedOutput: `ClusterBuilder "test-builder" patched (dry run)
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not patch a ClusterBuilder and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
spec:
  order:
  - group:
    - id: org.cloudfoundry.test-bp
  - group:
    - id: org.cloudfoundry.fake-bp
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-other-stack
  store:
    kind: ClusterStore
    name: some-other-store
  tag: some-other-tag
status:
  stack: {}
`

					testhelpers.CommandTest{
						KpackObjects: []runtime.Object{
							builder,
						},
						Args: []string{
							builder.Name,
							"--tag", "some-other-tag",
							"--stack", "some-other-stack",
							"--store", "some-other-store",
							"--order", "./testdata/patched-order.yaml",
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})
}
