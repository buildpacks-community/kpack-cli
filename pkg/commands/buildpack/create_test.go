// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack_test

import (
	"encoding/json"
	"testing"

	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/buildpack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

func TestBuildpackCreateCommand(t *testing.T) {
	spec.Run(t, "TestBuildpackCreateCommand", testCreateCommand(buildpack.NewCreateCommand))
}

func setLastAppliedAnnotation(b *buildv1alpha2.Buildpack) error {
	lastApplied, err := json.Marshal(b)
	if err != nil {
		return err
	}
	b.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = string(lastApplied)
	return nil
}

func testCreateCommand(cmd func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		const defaultNamespace = "some-default-namespace"
		var expectedBuildpack *buildv1alpha2.Buildpack

		it.Before(func() {
			expectedBuildpack = &buildv1alpha2.Buildpack{
				TypeMeta: metav1.TypeMeta{
					Kind:       buildv1alpha2.BuildpackKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-buildpack",
					Namespace:   "some-namespace",
					Annotations: map[string]string{},
				},
				Spec: buildv1alpha2.BuildpackSpec{
					ServiceAccountName: "default",
					ImageSource: corev1alpha1.ImageSource{
						Image: "some-registry.com/test-buildpack",
					},
				},
			}
		})

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
			return cmd(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("creates a Buildpack", func() {
			require.NoError(t, setLastAppliedAnnotation(expectedBuildpack))
			testhelpers.CommandTest{
				Args: []string{
					expectedBuildpack.Name,
					"--image", expectedBuildpack.Spec.Image,
					"-n", expectedBuildpack.Namespace,
				},
				ExpectedOutput: `Buildpack "test-buildpack" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuildpack,
				},
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("can creates a Buildpack with a custom service account", func() {
			expectedBuildpack.Spec.ServiceAccountName = "some-sa"
			require.NoError(t, setLastAppliedAnnotation(expectedBuildpack))
			testhelpers.CommandTest{
				Args: []string{
					expectedBuildpack.Name,
					"--image", expectedBuildpack.Spec.Image,
					"-n", expectedBuildpack.Namespace,
					"--service-account", "some-sa",
				},
				ExpectedOutput: `Buildpack "test-buildpack" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuildpack,
				},
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("creates a Buildpack with the default namespace", func() {
			expectedBuildpack.Namespace = defaultNamespace
			require.NoError(t, setLastAppliedAnnotation(expectedBuildpack))

			testhelpers.CommandTest{
				Args: []string{
					expectedBuildpack.Name,
					"--image", expectedBuildpack.Spec.Image,
				},
				ExpectedOutput: `Buildpack "test-buildpack" created
`,
				ExpectCreates: []runtime.Object{
					expectedBuildpack,
				},
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				require.NoError(t, setLastAppliedAnnotation(expectedBuildpack))
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Buildpack
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Buildpack","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-buildpack","namespace":"some-namespace","creationTimestamp":null},"spec":{"image":"some-registry.com/test-buildpack","serviceAccountName":"default"},"status":{}}'
  creationTimestamp: null
  name: test-buildpack
  namespace: some-namespace
spec:
  image: some-registry.com/test-buildpack
  serviceAccountName: default
status: {}
`

				testhelpers.CommandTest{
					Args: []string{
						expectedBuildpack.Name,
						"--image", expectedBuildpack.Spec.Image,
						"-n", expectedBuildpack.Namespace,
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectCreates: []runtime.Object{
						expectedBuildpack,
					},
				}.TestKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				require.NoError(t, setLastAppliedAnnotation(expectedBuildpack))
				const resourceJSON = `{
    "kind": "Buildpack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-buildpack",
        "namespace": "some-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Buildpack\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"test-buildpack\",\"namespace\":\"some-namespace\",\"creationTimestamp\":null},\"spec\":{\"image\":\"some-registry.com/test-buildpack\",\"serviceAccountName\":\"default\"},\"status\":{}}"
        }
    },
    "spec": {
        "image": "some-registry.com/test-buildpack",
        "serviceAccountName": "default"
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					Args: []string{
						expectedBuildpack.Name,
						"--image", expectedBuildpack.Spec.Image,
						"-n", expectedBuildpack.Namespace,
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectCreates: []runtime.Object{
						expectedBuildpack,
					},
				}.TestKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a Buildpack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Args: []string{
						expectedBuildpack.Name,
						"--image", expectedBuildpack.Spec.Image,
						"-n", expectedBuildpack.Namespace,
						"--dry-run",
					},
					ExpectedOutput: `Buildpack "test-buildpack" created (dry run)
`,
				}.TestKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("output flag is used", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Buildpack
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Buildpack","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-buildpack","namespace":"some-namespace","creationTimestamp":null},"spec":{"image":"some-registry.com/test-buildpack","serviceAccountName":"default"},"status":{}}'
  creationTimestamp: null
  name: test-buildpack
  namespace: some-namespace
spec:
  image: some-registry.com/test-buildpack
  serviceAccountName: default
status: {}
`

				it("does not create a Buildpack and prints the resource output", func() {
					testhelpers.CommandTest{
						Args: []string{
							expectedBuildpack.Name,
							"--image", expectedBuildpack.Spec.Image,
							"-n", expectedBuildpack.Namespace,
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
