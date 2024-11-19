// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	cbcmds "github.com/buildpacks-community/kpack-cli/pkg/commands/clusterbuilder"
	commandsfakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestClusterBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderPatchCommand", testPatchCommand(cbcmds.NewPatchCommand))
}

func testPatchCommand(clusterBuilderCommand func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		var (
			builder = &v1alpha2.ClusterBuilder{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha2.BuilderKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-builder",
				},
				Spec: v1alpha2.ClusterBuilderSpec{
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
					ServiceAccountRef: corev1.ObjectReference{
						Name:      "some-serviceaccount",
						Namespace: "some-namespace",
					},
				},
			}

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
		)

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
			return clusterBuilderCommand(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("patches a ClusterBuilder but does not update the default service account", func() {
			config.Data["default.repository.serviceaccount"] = "some-new-serviceaccount"
			config.Data["default.repository.serviceaccount.namespace"] = "some-new-namespace"
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					builder,
					config,
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
			}.TestKpack(t, cmdFunc)

			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("does not patch if there are no changes", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					builder,
					config,
				},
				Args: []string{
					builder.Name,
				},
				ExpectedOutput: `ClusterBuilder "test-builder" patched (no change)
`,
			}.TestKpack(t, cmdFunc)
		})

		it("patches a ClusterBuilder with buildpack flags", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					builder,
					config,
				},
				Args: []string{
					builder.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--buildpack", "org.cloudfoundry.test-bp",
					"--buildpack", "org.cloudfoundry.fake-bp@2.0.1",
				},
				ExpectedOutput: `ClusterBuilder "test-builder" patched
`,
				ExpectPatches: []string{
					`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"},{"id":"org.cloudfoundry.fake-bp","version":"2.0.1"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("returns error when buildpack and order flags are used together", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					builder,
					config,
				},
				Args: []string{
					builder.Name,
					"--order", "./testdata/patched-order.yaml",
					"--buildpack", "org.cloudfoundry.test-bp",
				},
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: cannot use --order and --buildpack together\n",
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Builder
metadata:
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
    namespace: some-namespace
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
					Objects: []runtime.Object{
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
				}.TestKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "Builder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-builder",
        "creationTimestamp": null
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
				}.TestKpack(t, cmdFunc)
			})

			when("there are no changes in the patch", func() {
				it("can output unpatched resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Builder
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
    name: some-serviceaccount
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

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							builder,
						},
						Args: []string{
							builder.Name,
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
			it("does not patch a ClusterBuilder and prints result with dry run indicated", func() {
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
					ExpectedOutput: `ClusterBuilder "test-builder" patched (dry run)
`,
				}.TestKpack(t, cmdFunc)
			})

			when("there are no changes in the patch", func() {
				it("does not patch and informs of no change", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							builder,
						},
						Args: []string{
							builder.Name,
							"--dry-run",
						},
						ExpectedOutput: `ClusterBuilder "test-builder" patched (dry run)
`,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not patch a ClusterBuilder and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Builder
metadata:
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
    namespace: some-namespace
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
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	}
}
