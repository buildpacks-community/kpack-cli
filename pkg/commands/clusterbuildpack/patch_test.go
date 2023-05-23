// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack_test

import (
	"testing"

	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
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
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuildpack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuildpackPatchCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackPatchCommand", testPatchCommand(clusterbuildpack.NewPatchCommand))
}

func testPatchCommand(cmd func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
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

			cbp = &buildv1alpha2.ClusterBuildpack{
				TypeMeta: metav1.TypeMeta{
					Kind:       buildv1alpha2.ClusterBuildpackKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-buildpack",
				},
				Spec: buildv1alpha2.ClusterBuildpackSpec{
					ImageSource: corev1alpha1.ImageSource{
						Image: "some-registry.com/test-buildpack",
					},
					ServiceAccountRef: &corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}
		)

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
			return cmd(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("patches a ClusterBuildpack but does not update the default service account", func() {
			config.Data["default.repository.serviceaccount"] = "some-new-serviceaccount"
			config.Data["default.repository.serviceaccount.namespace"] = "some-new-namespace"
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					cbp,
					config,
				},
				Args: []string{
					cbp.Name,
					"--image", "some-registry.com/some-other-buildpack",
				},
				ExpectedOutput: `Cluster Buildpack "test-buildpack" patched
`,
				ExpectPatches: []string{
					`{"spec":{"image":"some-registry.com/some-other-buildpack"}}`,
				},
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("does not patch if there are no changes", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					cbp,
					config,
				},
				Args: []string{
					cbp.Name,
					"--image", cbp.Spec.Image,
				},
				ExpectedOutput: `Cluster Buildpack "test-buildpack" patched (no change)
`,
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterBuildpack
metadata:
  creationTimestamp: null
  name: test-buildpack
spec:
  image: some-registry.com/some-other-buildpack
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						cbp,
					},
					Args: []string{
						cbp.Name,
						"--image", "some-registry.com/some-other-buildpack",
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
    "kind": "ClusterBuildpack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-buildpack",
        "creationTimestamp": null
    },
    "spec": {
        "image": "some-registry.com/some-other-buildpack",
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						cbp,
					},
					Args: []string{
						cbp.Name,
						"--image", "some-registry.com/some-other-buildpack",
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
kind: ClusterBuildpack
metadata:
  creationTimestamp: null
  name: test-buildpack
spec:
  image: some-registry.com/test-buildpack
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							cbp,
						},
						Args: []string{
							cbp.Name,
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a ClusterBuildpack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						cbp,
					},
					Args: []string{
						cbp.Name,
						"--image", "some-registry.com/some-other-buildpack",
						"--dry-run",
					},
					ExpectedOutput: `Cluster Buildpack "test-buildpack" patched (dry run)
`,
				}.TestKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("there are no changes in the patch", func() {
				it("does not patch and informs of no change", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							cbp,
						},
						Args: []string{
							cbp.Name,
							"--dry-run",
						},
						ExpectedOutput: `Cluster Buildpack "test-buildpack" patched (dry run)
`,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not create a ClusterBuildpack and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterBuildpack
metadata:
  creationTimestamp: null
  name: test-buildpack
spec:
  image: some-registry.com/some-other-buildpack
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							cbp,
						},
						Args: []string{
							cbp.Name,
							"--image", "some-registry.com/some-other-buildpack",
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
