// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/builder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuilderSaveCommand(t *testing.T) {
	spec.Run(t, "TestBuilderSaveCommand", testBuilderSaveCommand)
}

func testBuilderSaveCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	var (
		expectedBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.BuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-builder",
				Namespace: "some-namespace",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Builder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccount":"default"},"status":{"stack":{}}}`,
				},
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder",
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
				ServiceAccount: "default",
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return builder.NewSaveCommand(clientSetProvider)
	}

	when ("creating", func() {
		it("creates a Builder when it does not exist", func() {
			testhelpers.CommandTest{
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
					"-n", expectedBuilder.Namespace,
				},
				ExpectedOutput: `"test-builder" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("creates a Builder with the default namespace, store, and stack", func() {
			expectedBuilder.Namespace = defaultNamespace
			expectedBuilder.Spec.Stack.Name = "default"
			expectedBuilder.Spec.Store.Name = "default"
			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"Builder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"default"},"store":{"kind":"ClusterStore","name":"default"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccount":"default"},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--order", "./testdata/order.yaml",
				},
				ExpectedOutput: "\"test-builder\" created\n",
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
		})

		when("dry run is specified", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: Builder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Builder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccount":"default"},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
  namespace: some-namespace
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccount: default
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: some-registry.com/test-builder
status:
  stack: {}
`
			const resourceJSON = `{
    "kind": "Builder",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-builder",
        "namespace": "some-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Builder\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"test-builder\",\"namespace\":\"some-namespace\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry.com/test-builder\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"org.cloudfoundry.nodejs\"}]},{\"group\":[{\"id\":\"org.cloudfoundry.go\"}]}],\"serviceAccount\":\"default\"},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "some-registry.com/test-builder",
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
        "serviceAccount": "default"
    },
    "status": {
        "stack": {}
    }
}
`

			when("without an output format", func() {
				it("does not create a Builder and defaults resource output to yaml format", func() {
					testhelpers.CommandTest{
						Args: []string{
							expectedBuilder.Name,
							"--tag", expectedBuilder.Spec.Tag,
							"--stack", expectedBuilder.Spec.Stack.Name,
							"--store", expectedBuilder.Spec.Store.Name,
							"--order", "./testdata/order.yaml",
							"-n", expectedBuilder.Namespace,
							"--dry-run",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
				})
			})

			it("does not create a Builder and outputs the resource in yaml format", func() {
				testhelpers.CommandTest{
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"-n", expectedBuilder.Namespace,
						"--dry-run", "--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
			})

			it("does not create a Builder and outputs the resource in json format", func() {
				testhelpers.CommandTest{
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"-n", expectedBuilder.Namespace,
						"--dry-run", "--output", "json",
					},
					ExpectedOutput: resourceJSON,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when ("patching", func() {
		it("patches a Builder when it already exists", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					expectedBuilder,
				},
				Args: []string{
					expectedBuilder.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
					"-n", expectedBuilder.Namespace,
				},
				ExpectedOutput: "\"test-builder\" patched\n",
				ExpectPatches: []string{
					`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		when("dry run is specified", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: Builder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Builder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccount":"default"},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
  namespace: some-namespace
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccount: default
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: some-registry.com/test-builder
status:
  stack: {}
`
			const resourceJSON = `{
    "kind": "Builder",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-builder",
        "namespace": "some-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Builder\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"test-builder\",\"namespace\":\"some-namespace\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry.com/test-builder\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"org.cloudfoundry.nodejs\"}]},{\"group\":[{\"id\":\"org.cloudfoundry.go\"}]}],\"serviceAccount\":\"default\"},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "some-registry.com/test-builder",
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
        "serviceAccount": "default"
    },
    "status": {
        "stack": {}
    }
}
`

			it("does not patch the Builder and outputs the resource in yaml format", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						expectedBuilder,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", "some-other-tag",
						"--stack", "some-other-stack",
						"--store", "some-other-store",
						"--order", "./testdata/patched-order.yaml",
						"-n", expectedBuilder.Namespace,
						"--dry-run", "--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
			})

			it("does not patch the Builder and outputs the resource in json format", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						expectedBuilder,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", "some-other-tag",
						"--stack", "some-other-stack",
						"--store", "some-other-store",
						"--order", "./testdata/patched-order.yaml",
						"-n", expectedBuilder.Namespace,
						"--dry-run", "--output", "json",
					},
					ExpectedOutput: resourceJSON,
				}.TestKpack(t, cmdFunc)
			})

			when("without an output format", func() {
				it("does not patch the Builder and defaults resource output to yaml format", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							expectedBuilder,
						},
						Args: []string{
							expectedBuilder.Name,
							"--tag", "some-other-tag",
							"--stack", "some-other-stack",
							"--store", "some-other-store",
							"--order", "./testdata/patched-order.yaml",
							"-n", expectedBuilder.Namespace,
							"--dry-run",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("without any changes", func() {
				it("does not patch and informs user nothing to patch", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							expectedBuilder,
						},
						Args: []string{
							expectedBuilder.Name,
							"-n", expectedBuilder.Namespace,
							"--dry-run", "--output", "yaml",
						},
						ExpectedOutput: "nothing to patch\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}
