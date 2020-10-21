// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderCreateCommand(t *testing.T) {
	spec.Run(t, "TestBuilderCreateCommand", testClusterBuilderCreateCommand)
}

func testClusterBuilderCreateCommand(t *testing.T, when spec.G, it spec.S) {
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

		expectedBuilder = &v1alpha1.ClusterBuilder{
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

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterbuilder.NewCreateCommand(clientSetProvider)
	}

	it("creates a ClusterBuilder", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
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
	})

	it("creates a ClusterBuilder with the default stack", func() {
		expectedBuilder.Spec.Stack.Name = "default"
		expectedBuilder.Spec.Store.Name = "default"
		expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"default"},"store":{"kind":"ClusterStore","name":"default"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
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

	it("creates a ClusterBuilder with the canonical tag when tag is not specified", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
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

	it("fails when kp-config map is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
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
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
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
				expectedBuilder.Name,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
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
				K8sObjects: []runtime.Object{
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
			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.go"},{"id":"org.cloudfoundry.nodejs","version":"1"},{"id":"org.cloudfoundry.ruby","version":"1.2.3"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
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
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedBuilder.Name,
						"--tag", expectedBuilder.Spec.Tag,
						"--order", "./testdata/order.yaml",
						"--buildpack", "some-buildpack-name",
					},
					ExpectErr: true,
					ExpectedOutput: `Error: cannot use --order and --buildpack together
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

}
