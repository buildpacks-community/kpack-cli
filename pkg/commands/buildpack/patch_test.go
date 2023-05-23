// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack_test

import (
	"testing"

	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/buildpack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuildpackPatchCommand(t *testing.T) {
	spec.Run(t, "TestBuildpackPatchCommand", testPatchCommand(buildpack.NewPatchCommand))
}

func testPatchCommand(cmd func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		const defaultNamespace = "some-default-namespace"

		var (
			bp = &buildv1alpha2.Buildpack{
				TypeMeta: metav1.TypeMeta{
					Kind:       buildv1alpha2.BuildpackKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-buildpack",
					Namespace: "some-namespace",
				},
				Spec: buildv1alpha2.BuildpackSpec{
					ServiceAccountName: "default",
					ImageSource: corev1alpha1.ImageSource{
						Image: "some-registry.com/test-buildpack",
					},
				},
			}
		)

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
			return cmd(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("patches a Buildpack", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					bp,
				},
				Args: []string{
					bp.Name,
					"--image", "some-registry.com/some-other-buildpack",
					"-n", bp.Namespace,
					"--service-account", "some-other-sa",
				},
				ExpectedOutput: `Buildpack "test-buildpack" patched
`,
				ExpectPatches: []string{
					`{"spec":{"image":"some-registry.com/some-other-buildpack","serviceAccountName":"some-other-sa"}}`,
				},
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("patches a Buildpack in the default namespace", func() {
			bp.Namespace = defaultNamespace

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					bp,
				},
				Args: []string{
					bp.Name,
					"--image", "some-registry.com/some-other-buildpack",
					"--service-account", "some-other-sa",
				},
				ExpectedOutput: `Buildpack "test-buildpack" patched
`,
				ExpectPatches: []string{
					`{"spec":{"image":"some-registry.com/some-other-buildpack","serviceAccountName":"some-other-sa"}}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("does not patch if there are no changes", func() {
			bp.Spec.ServiceAccountName = "some-other-sa"
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					bp,
				},
				Args: []string{
					bp.Name,
					"-n", bp.Namespace,
				},
				ExpectedOutput: `Buildpack "test-buildpack" patched (no change)
`,
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Buildpack
metadata:
  creationTimestamp: null
  name: test-buildpack
  namespace: some-namespace
spec:
  image: some-registry.com/some-other-buildpack
  serviceAccountName: default
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						bp,
					},
					Args: []string{
						bp.Name,
						"--image", "some-registry.com/some-other-buildpack",
						"-n", bp.Namespace,
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectPatches: []string{
						`{"spec":{"image":"some-registry.com/some-other-buildpack"}}`,
					},
				}.TestKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "Buildpack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-buildpack",
        "namespace": "some-namespace",
        "creationTimestamp": null
    },
    "spec": {
        "image": "some-registry.com/some-other-buildpack",
        "serviceAccountName": "default"
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						bp,
					},
					Args: []string{
						bp.Name,
						"--image", "some-registry.com/some-other-buildpack",
						"-n", bp.Namespace,
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectPatches: []string{
						`{"spec":{"image":"some-registry.com/some-other-buildpack"}}`,
					},
				}.TestKpack(t, cmdFunc)
			})

			when("there are no changes in the patch", func() {
				it("can output unpatched resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Buildpack
metadata:
  creationTimestamp: null
  name: test-buildpack
  namespace: some-namespace
spec:
  image: some-registry.com/test-buildpack
  serviceAccountName: default
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							bp,
						},
						Args: []string{
							bp.Name,
							"-n", bp.Namespace,
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a Buildpack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						bp,
					},
					Args: []string{
						bp.Name,
						"--image", "some-registry.com/some-other-buildpack",
						"-n", bp.Namespace,
						"--dry-run",
					},
					ExpectedOutput: `Buildpack "test-buildpack" patched (dry run)
`,
				}.TestKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("there are no changes in the patch", func() {
				it("does not patch and informs of no change", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							bp,
						},
						Args: []string{
							bp.Name,
							"-n", bp.Namespace,
							"--dry-run",
						},
						ExpectedOutput: `Buildpack "test-buildpack" patched (dry run)
`,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not create a Buildpack and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Buildpack
metadata:
  creationTimestamp: null
  name: test-buildpack
  namespace: some-namespace
spec:
  image: some-registry.com/some-other-buildpack
  serviceAccountName: default
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							bp,
						},
						Args: []string{
							bp.Name,
							"--image", "some-registry.com/some-other-buildpack",
							"-n", bp.Namespace,
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
