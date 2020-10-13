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

func TestBuilderCreateCommand(t *testing.T) {
	spec.Run(t, "TestBuilderCreateCommand", testBuilderCreateCommand)
}

func testBuilderCreateCommand(t *testing.T, when spec.G, it spec.S) {
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
		return builder.NewCreateCommand(clientSetProvider)
	}

	it("creates a Builder", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
				"--order", "./testdata/order.yaml",
				"-n", expectedBuilder.Namespace,
			},
			ExpectedOutput: `Builder "test-builder" created
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
			ExpectedOutput: `Builder "test-builder" created
`,
			ExpectCreates: []runtime.Object{
				expectedBuilder,
			},
		}.TestKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
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

			testhelpers.CommandTest{
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
					"-n", expectedBuilder.Namespace,
					"--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
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

			testhelpers.CommandTest{
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
					"-n", expectedBuilder.Namespace,
					"--output", "json",
				},
				ExpectedOutput: resourceJSON,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not create a Builder and prints result with dry run indicated", func() {
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
				ExpectedOutput: `Builder "test-builder" created (dry run)
`,
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
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

			it("does not create a Builder and prints the resource output", func() {
				testhelpers.CommandTest{
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--stack", expectedBuilder.Spec.Stack.Name,
						"--store", expectedBuilder.Spec.Store.Name,
						"--order", "./testdata/order.yaml",
						"-n", expectedBuilder.Namespace,
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("buildpack flag is used", func() {
		it("creates a builder using the buildpack flag", func() {

			expectedBuilder.Spec.Order = []v1alpha1.OrderEntry{
				{
					Group: []v1alpha1.BuildpackRef{
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id: "org.cloudfoundry.go",
							},
						},
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id:      "org.cloudfoundry.nodejs",
								Version: "1",
							},
						},
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id:      "org.cloudfoundry.ruby",
								Version: "1.2.3",
							},
						},
					},
				},
			}
			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"Builder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.go"},{"id":"org.cloudfoundry.nodejs","version":"1"},{"id":"org.cloudfoundry.ruby","version":"1.2.3"}]}],"serviceAccount":"default"},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--buildpack", "org.cloudfoundry.go,org.cloudfoundry.nodejs@1",
					"--buildpack", "org.cloudfoundry.ruby@1.2.3",
					"-n", expectedBuilder.Namespace,
				},
				ExpectedOutput: `Builder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
		})

		when("buildpack and order flags are used together", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--order", "./testdata/order.yaml",
						"--buildpack", "some-buildpack-name",
					},
					ExpectErr: true,
					ExpectedOutput: `Error: cannot use --order and --buildpack together
`,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

}
