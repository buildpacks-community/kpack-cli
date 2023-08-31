// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"encoding/json"
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	buildercmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/builder"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuilderCreateCommand(t *testing.T) {
	spec.Run(t, "TestBuilderCreateCommand", testCreateCommand(buildercmds.NewCreateCommand))
}

func setLastAppliedAnnotation(b *v1alpha2.Builder) error {
	lastApplied, err := json.Marshal(b)
	if err != nil {
		return err
	}
	b.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = string(lastApplied)
	return nil
}

func testCreateCommand(builderCommand func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		const defaultNamespace = "some-default-namespace"
		var expectedBuilder *v1alpha2.Builder
		it.Before(func() {
			expectedBuilder = &v1alpha2.Builder{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha2.BuilderKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:        "test-builder",
					Namespace:   "some-namespace",
					Annotations: map[string]string{},
				},
				Spec: v1alpha2.NamespacedBuilderSpec{
					BuilderSpec: v1alpha2.BuilderSpec{
						Tag: "some-registry.com/test-builder",
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
					ServiceAccountName: "default",
				},
			}
		})

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
			return builderCommand(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("creates a Builder", func() {
			require.NoError(t, setLastAppliedAnnotation(expectedBuilder))
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
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("can creates a Builder with a custom service account", func() {
			expectedBuilder.Spec.ServiceAccountName = "some-sa"
			require.NoError(t, setLastAppliedAnnotation(expectedBuilder))
			testhelpers.CommandTest{
				Args: []string{
					expectedBuilder.Name,
					"--tag", expectedBuilder.Spec.Tag,
					"--stack", expectedBuilder.Spec.Stack.Name,
					"--store", expectedBuilder.Spec.Store.Name,
					"--order", "./testdata/order.yaml",
					"-n", expectedBuilder.Namespace,
					"--service-account", "some-sa",
				},
				ExpectedOutput: `Builder "test-builder" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("creates a Builder with the default namespace and stack, store is not set", func() {
			expectedBuilder.Namespace = defaultNamespace
			expectedBuilder.Spec.Stack.Name = "default"
			expectedBuilder.Spec.Store = corev1.ObjectReference{}
			require.NoError(t, setLastAppliedAnnotation(expectedBuilder))

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
				require.NoError(t, setLastAppliedAnnotation(expectedBuilder))
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Builder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Builder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountName":"default"},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
  namespace: some-namespace
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccountName: default
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
				require.NoError(t, setLastAppliedAnnotation(expectedBuilder))
				const resourceJSON = `{
    "kind": "Builder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-builder",
        "namespace": "some-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Builder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"test-builder\",\"namespace\":\"some-namespace\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry.com/test-builder\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"org.cloudfoundry.nodejs\"}]},{\"group\":[{\"id\":\"org.cloudfoundry.go\"}]}],\"serviceAccountName\":\"default\"},\"status\":{\"stack\":{}}}"
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
        "serviceAccountName": "default"
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
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("output flag is used", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Builder
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Builder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-builder","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountName":"default"},"status":{"stack":{}}}'
  creationTimestamp: null
  name: test-builder
  namespace: some-namespace
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccountName: default
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
				require.NoError(t, setLastAppliedAnnotation(expectedBuilder))

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
						ExpectedErrorOutput: `Error: cannot use --order and --buildpack together
`,
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	}
}
