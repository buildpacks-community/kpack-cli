// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack_test

import (
	"encoding/json"
	"testing"

	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuildpack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuildpackCreateCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackCreateCommand", testCreateCommand(clusterbuildpack.NewCreateCommand))
}

func setLastAppliedAnnotation(b *buildv1alpha2.ClusterBuildpack) error {
	lastApplied, err := json.Marshal(b)
	if err != nil {
		return err
	}
	b.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = string(lastApplied)
	return nil
}

func testCreateCommand(cmd func(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		var (
			expectedClusterBuildpack *buildv1alpha2.ClusterBuildpack

			config = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository.serviceaccount":           "some-serviceaccount",
					"default.repository.serviceaccount.namespace": "some-namespace",
				},
			}
		)

		it.Before(func() {
			expectedClusterBuildpack = &buildv1alpha2.ClusterBuildpack{
				TypeMeta: metav1.TypeMeta{
					Kind:       buildv1alpha2.ClusterBuildpackKind,
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-buildpack",
					Annotations: map[string]string{},
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
		})

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(k8sClientset *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientset, kpackClientSet)
			return cmd(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("creates a ClusterBuildpack", func() {
			require.NoError(t, setLastAppliedAnnotation(expectedClusterBuildpack))
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					expectedClusterBuildpack.Name,
					"--image", expectedClusterBuildpack.Spec.Image,
				},
				ExpectedOutput: `Cluster Buildpack "test-buildpack" created
`,
				ExpectCreates: []runtime.Object{
					expectedClusterBuildpack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				require.NoError(t, setLastAppliedAnnotation(expectedClusterBuildpack))
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterBuildpack
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuildpack","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-buildpack","creationTimestamp":null},"spec":{"image":"some-registry.com/test-buildpack","serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{}}'
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
						config,
					},
					Args: []string{
						expectedClusterBuildpack.Name,
						"--image", expectedClusterBuildpack.Spec.Image,
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectCreates: []runtime.Object{
						expectedClusterBuildpack,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				require.NoError(t, setLastAppliedAnnotation(expectedClusterBuildpack))
				const resourceJSON = `{
    "kind": "ClusterBuildpack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "test-buildpack",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuildpack\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"test-buildpack\",\"creationTimestamp\":null},\"spec\":{\"image\":\"some-registry.com/test-buildpack\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{}}"
        }
    },
    "spec": {
        "image": "some-registry.com/test-buildpack",
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
						config,
					},
					Args: []string{
						expectedClusterBuildpack.Name,
						"--image", expectedClusterBuildpack.Spec.Image,
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectCreates: []runtime.Object{
						expectedClusterBuildpack,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a ClusterBuildpack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedClusterBuildpack.Name,
						"--image", expectedClusterBuildpack.Spec.Image,
						"--dry-run",
					},
					ExpectedOutput: `Cluster Buildpack "test-buildpack" created (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("output flag is used", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterBuildpack
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuildpack","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"test-buildpack","creationTimestamp":null},"spec":{"image":"some-registry.com/test-buildpack","serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{}}'
  creationTimestamp: null
  name: test-buildpack
spec:
  image: some-registry.com/test-buildpack
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status: {}
`

				it("does not create a ClusterBuildpack and prints the resource output", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
						},
						Args: []string{
							expectedClusterBuildpack.Name,
							"--image", expectedClusterBuildpack.Spec.Image,
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	}
}
