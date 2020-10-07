// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	clusterstackcmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCreateCommand(t *testing.T) {
	spec.Run(t, "TestCreateCommand", testCreateCommand)
}

func testCreateCommand(t *testing.T, when spec.G, it spec.S) {
	buildImage, buildImageId, runImage, runImageId := makeStackImages(t, "some-stack-id")

	fetcher := &fakes.Fetcher{}
	fetcher.AddImage("some-build-image", buildImage)
	fetcher.AddImage("some-run-image", runImage)

	relocator := &fakes.Relocator{}

	stackFactory := &clusterstack.Factory{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "some-registry.io/some-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterstackcmds.NewCreateCommand(clientSetProvider, stackFactory)
	}

	it("creates a stack", func() {
		expectedStack := &v1alpha1.ClusterStack{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterStackKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-stack",
			},
			Spec: v1alpha1.ClusterStackSpec{
				Id: "some-stack-id",
				BuildImage: v1alpha1.ClusterStackSpecImage{
					Image: "some-registry.io/some-repo/build@" + buildImageId,
				},
				RunImage: v1alpha1.ClusterStackSpecImage{
					Image: "some-registry.io/some-repo/run@" + runImageId,
				},
			},
		}

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectedOutput: `Creating Cluster Stack...
Uploading to 'some-registry.io/some-repo'...
"some-stack" created
`,
			ExpectCreates: []runtime.Object{
				expectedStack,
			},
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when canonical.repository key is not found in kp-config configmap", func() {
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
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("validates build stack ID is equal to run stack ID", func() {
		_, _, runImage, _ := makeStackImages(t, "some-other-stack-id")

		fetcher.AddImage("some-other-run-image", runImage)

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-other-run-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Creating Cluster Stack...
Error: build stack 'some-stack-id' does not match run stack 'some-other-stack-id'
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		expectedStack := &v1alpha1.ClusterStack{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterStackKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-stack",
			},
			Spec: v1alpha1.ClusterStackSpec{
				Id: "some-stack-id",
				BuildImage: v1alpha1.ClusterStackSpecImage{
					Image: "some-registry.io/some-repo/build@" + buildImageId,
				},
				RunImage: v1alpha1.ClusterStackSpecImage{
					Image: "some-registry.io/some-repo/run@" + runImageId,
				},
			},
		}

		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-registry.io/some-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-stack-id
  runImage:
    image: some-registry.io/some-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
status:
  buildImage: {}
  runImage: {}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
					"--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Creating Cluster Stack...
Uploading to 'some-registry.io/some-repo'...
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-stack",
        "creationTimestamp": null
    },
    "spec": {
        "id": "some-stack-id",
        "buildImage": {
            "image": "some-registry.io/some-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d"
        },
        "runImage": {
            "image": "some-registry.io/some-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
					"--output", "json",
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Creating Cluster Stack...
Uploading to 'some-registry.io/some-repo'...
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not create a clusterstack and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
					"--dry-run",
				},
				ExpectedOutput: `Creating Cluster Stack... (dry run)
Uploading to 'some-registry.io/some-repo'...
"some-stack" created (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not create a clusterstack and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-registry.io/some-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-stack-id
  runImage:
    image: some-registry.io/some-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
status:
  buildImage: {}
  runImage: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"some-stack",
						"--build-image", "some-build-image",
						"--run-image", "some-run-image",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating Cluster Stack... (dry run)
Uploading to 'some-registry.io/some-repo'...
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}
