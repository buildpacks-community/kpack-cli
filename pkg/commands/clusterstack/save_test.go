// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	clientgotesting "k8s.io/client-go/testing"

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

func TestSaveCommand(t *testing.T) {
	spec.Run(t, "TestSaveCommand", testSaveCommand)
}

func testSaveCommand(t *testing.T, when spec.G, it spec.S) {
	fetcher := &fakes.Fetcher{}

	buildImage, buildImageId, runImage, runImageId := makeStackImages(t, "some-stack-id")
	fetcher.AddImage("some-build-image", buildImage)
	fetcher.AddImage("some-run-image", runImage)

	newBuildImage, newBuildImageId, newRunImage, newRunImageId := makeStackImages(t, "some-new-id")
	fetcher.AddImage("some-new-build-image", newBuildImage)
	fetcher.AddImage("some-new-run-image", newRunImage)

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
		return clusterstackcmds.NewSaveCommand(clientSetProvider, stackFactory)
	}

	when("creating", func() {
		it("creates a stack when it does not exist", func() {
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
				},
				ExpectedOutput: "Uploading to 'some-registry.io/some-repo'...\n\"some-stack\" created\n",
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
				ExpectErr:      true,
				ExpectedOutput: "Error: build stack 'some-stack-id' does not match run stack 'some-other-stack-id'\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("updating", func() {
		it("updates the stack id, run image, and build image when the clusterstack does exist", func() {
			stack := &v1alpha1.ClusterStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-stack",
				},
				Spec: v1alpha1.ClusterStackSpec{
					Id: "some-old-id",
					BuildImage: v1alpha1.ClusterStackSpecImage{
						Image: "some-old-build-image",
					},
					RunImage: v1alpha1.ClusterStackSpecImage{
						Image: "some-old-run-image",
					},
				},
				Status: v1alpha1.ClusterStackStatus{
					ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
						Id: "some-old-id",
						BuildImage: v1alpha1.ClusterStackStatusImage{
							LatestImage: "some-registry.io/old-repo/build@" + buildImageId,
							Image:       "some-old-build-image",
						},
						RunImage: v1alpha1.ClusterStackStatusImage{
							LatestImage: "some-registry.io/old-repo/run@" + runImageId,
							Image:       "some-old-run-image",
						},
					},
				},
			}

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
				ExpectErr: false,
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStack{
							ObjectMeta: stack.ObjectMeta,
							Spec: v1alpha1.ClusterStackSpec{
								Id: "some-new-id",
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.io/some-repo/build@" + newBuildImageId,
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.io/some-repo/run@" + newRunImageId,
								},
							},
							Status: stack.Status,
						},
					},
				},
				ExpectedOutput: "Uploading to 'some-registry.io/some-repo'...\nClusterStack \"some-stack\" Updated\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})
}
