// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/builder"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestBuilderPatchCommand", testBuilderPatchCommand)
}

func testBuilderPatchCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	var (
		bldr = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder",
				Namespace: "some-namespace",
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
					Order: []corev1alpha1.OrderEntry{
						{
							Group: []corev1alpha1.BuildpackRef{
								{
									BuildpackInfo: corev1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []corev1alpha1.BuildpackRef{
								{
									BuildpackInfo: corev1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountName: "default",
			},
		}
	)

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return builder.NewPatchCommand(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	it("patches a Builder", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"--tag", "some-other-tag",
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
				"-n", bldr.Namespace,
			},
			ExpectedOutput: `Builder "test-builder" patched
`,
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
			},
		}.TestKpack(t, cmdFunc)
		require.Len(t, fakeWaiter.WaitCalls, 1)
	})

	it("patches a Builder in the default namespace", func() {
		bldr.Namespace = defaultNamespace

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"--tag", "some-other-tag",
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
			},
			ExpectedOutput: `Builder "test-builder" patched
`,
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("does not patch if there are no changes", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"-n", bldr.Namespace,
			},
			ExpectedOutput: `Builder "test-builder" patched (no change)
`,
		}.TestKpack(t, cmdFunc)
	})

	it("patches a Builder with buildpack flags", func() {
		bldr.Namespace = defaultNamespace

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"--tag", "some-other-tag",
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--buildpack", "org.cloudfoundry.test-bp",
				"--buildpack", "org.cloudfoundry.fake-bp@2.0.1",
			},
			ExpectedOutput: `Builder "test-builder" patched
`,
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"},{"id":"org.cloudfoundry.fake-bp","version":"2.0.1"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("returns error when buildpack and order flags are used together", func() {
		bldr.Namespace = defaultNamespace

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
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
  namespace: some-namespace
spec:
  order:
  - group:
    - id: org.cloudfoundry.test-bp
  - group:
    - id: org.cloudfoundry.fake-bp
  serviceAccountName: default
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
					bldr,
				},
				Args: []string{
					bldr.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
					"-n", bldr.Namespace,
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
        "namespace": "some-namespace",
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
        "serviceAccountName": "default"
    },
    "status": {
        "stack": {}
    }
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					bldr,
				},
				Args: []string{
					bldr.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
					"-n", bldr.Namespace,
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
					Objects: []runtime.Object{
						bldr,
					},
					Args: []string{
						bldr.Name,
						"-n", bldr.Namespace,
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run flag is used", func() {
		it("does not create a Builder and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					bldr,
				},
				Args: []string{
					bldr.Name,
					"--tag", "some-other-tag",
					"--stack", "some-other-stack",
					"--store", "some-other-store",
					"--order", "./testdata/patched-order.yaml",
					"-n", bldr.Namespace,
					"--dry-run",
				},
				ExpectedOutput: `Builder "test-builder" patched (dry run)
`,
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 0)
		})

		when("there are no changes in the patch", func() {
			it("does not patch and informs of no change", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						bldr,
					},
					Args: []string{
						bldr.Name,
						"-n", bldr.Namespace,
						"--dry-run",
					},
					ExpectedOutput: `Builder "test-builder" patched (dry run)
`,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("output flag is used", func() {
			it("does not create a Builder and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Builder
metadata:
  creationTimestamp: null
  name: test-builder
  namespace: some-namespace
spec:
  order:
  - group:
    - id: org.cloudfoundry.test-bp
  - group:
    - id: org.cloudfoundry.fake-bp
  serviceAccountName: default
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
						bldr,
					},
					Args: []string{
						bldr.Name,
						"--tag", "some-other-tag",
						"--stack", "some-other-stack",
						"--store", "some-other-store",
						"--order", "./testdata/patched-order.yaml",
						"-n", bldr.Namespace,
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}
