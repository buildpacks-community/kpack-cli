// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderPatchCommand", testClusterBuilderPatchCommand)
}

func testClusterBuilderPatchCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		builder = &v1alpha1.ClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder",
			},
			Spec: v1alpha1.ClusterBuilderSpec{
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
				ServiceAccountRef: corev1.ObjectReference{
					Name:      "some-service-account",
					Namespace: "some-namespace",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuilder.NewPatchCommand(clientSetProvider)
	}

	it("patches a ClusterBuilder", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
				"--tag", "some-other-tag",
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("does not patch if there are no changes", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
			},
			ExpectedOutput: "nothing to patch\n",
		}.TestKpack(t, cmdFunc)
	})

	when("dry run is specified", func() {
		const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  creationTimestamp: null
  name: test-builder
spec:
  order:
  - group:
    - id: org.cloudfoundry.nodejs
  - group:
    - id: org.cloudfoundry.go
  serviceAccountRef:
    name: some-service-account
    namespace: some-namespace
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
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-builder",
        "creationTimestamp": null
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
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-service-account"
        }
    },
    "status": {
        "stack": {}
    }
}
`

		it("does not patch a ClusterBuilder and outputs the resource in yaml format", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					builder,
				},
				Args: []string{
					builder.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
					"--dry-run", "--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
			}.TestKpack(t, cmdFunc)
		})

		it("does not patch a ClusterBuilder and outputs the resource in json format", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					builder,
				},
				Args: []string{
					builder.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
					"--dry-run", "--output", "json",
				},
				ExpectedOutput: resourceJSON,
			}.TestKpack(t, cmdFunc)
		})

		when("without an output format", func() {
			it("does not patch a ClusterBuilder and defaults resource output to yaml format", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
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
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("without any changes", func() {
			it("does not patch and informs user nothing to patch", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						builder,
					},
					Args: []string{
						builder.Name,
						"--dry-run", "--output", "yaml",
					},
					ExpectedOutput: "nothing to patch\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}
