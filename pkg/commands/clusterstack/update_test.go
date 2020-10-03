// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	clstrstkcmd "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand)
}

func testUpdateCommand(t *testing.T, when spec.G, it spec.S) {
	fetcher := &fakes.Fetcher{}

	oldBuildImage, oldBuildImageId, oldRunImage, oldRunImageId := makeStackImages(t, "some-old-id")
	fetcher.AddImage("some-old-build-image", oldBuildImage)
	fetcher.AddImage("some-old-run-image", oldRunImage)

	newBuildImage, newBuildImageId, newRunImage, newRunImageId := makeStackImages(t, "some-new-id")
	fetcher.AddImage("some-new-build-image", newBuildImage)
	fetcher.AddImage("some-new-run-image", newRunImage)

	relocator := &fakes.Relocator{}

	stackFactory := &clusterstack.Factory{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

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
					LatestImage: "some-registry.com/old-repo/build@" + oldBuildImageId,
					Image:       "some-old-build-image",
				},
				RunImage: v1alpha1.ClusterStackStatusImage{
					LatestImage: "some-registry.com/old-repo/run@" + oldRunImageId,
					Image:       "some-old-run-image",
				},
			},
		},
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "some-registry.com/some-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clstrstkcmd.NewUpdateCommand(clientSetProvider, stackFactory)
	}

	it("updates the stack id, run image, and build image", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				stack,
			},
			Args: []string{"some-stack",
				"--build-image", "some-new-build-image",
				"--run-image", "some-new-run-image",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &v1alpha1.ClusterStack{
						ObjectMeta: stack.ObjectMeta,
						Spec: v1alpha1.ClusterStackSpec{
							Id: "some-new-id",
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: "some-registry.com/some-repo/build@" + newBuildImageId,
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: "some-registry.com/some-repo/run@" + newRunImageId,
							},
						},
						Status: stack.Status,
					},
				},
			},
			ExpectedOutput: "Uploading to 'some-registry.com/some-repo'...\nClusterStack \"some-stack\" Updated\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("does not add stack images with the same digest", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:           []string{"some-stack", "--build-image", "some-old-build-image", "--run-image", "some-old-run-image"},
			ExpectErr:      false,
			ExpectedOutput: "Uploading to 'some-registry.com/some-repo'...\nBuild and Run images already exist in stack\nClusterStack Unchanged\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("returns error when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("returns error when canonical.repository key is not found in kp-config configmap", func() {
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
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("returns error when build image and run image have different stack Ids", func() {
		_, _, runImage, _ := makeStackImages(t, "other-stack-id")

		fetcher.AddImage("some-new-run-image", runImage)

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:           []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr:      true,
			ExpectedOutput: "Error: build stack 'some-new-id' does not match run stack 'other-stack-id'\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-registry.com/some-repo/build@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
  id: some-new-id
  runImage:
    image: some-registry.com/some-repo/run@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
status:
  buildImage:
    image: some-old-build-image
    latestImage: some-registry.com/old-repo/build@sha256:f845e3c8d069d56623cbd2d98e811bb63c81bfa0b51d86fe2e51f322046f38b6
  id: some-old-id
  runImage:
    image: some-old-run-image
    latestImage: some-registry.com/old-repo/run@sha256:f845e3c8d069d56623cbd2d98e811bb63c81bfa0b51d86fe2e51f322046f38b6
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-new-build-image",
					"--run-image", "some-new-run-image",
					"--output", "yaml",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStack{
							ObjectMeta: stack.ObjectMeta,
							Spec: v1alpha1.ClusterStackSpec{
								Id: "some-new-id",
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.com/some-repo/build@" + newBuildImageId,
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.com/some-repo/run@" + newRunImageId,
								},
							},
							Status: stack.Status,
						},
					},
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Uploading to 'some-registry.com/some-repo'...
`,
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
        "id": "some-new-id",
        "buildImage": {
            "image": "some-registry.com/some-repo/build@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7"
        },
        "runImage": {
            "image": "some-registry.com/some-repo/run@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7"
        }
    },
    "status": {
        "id": "some-old-id",
        "buildImage": {
            "latestImage": "some-registry.com/old-repo/build@sha256:f845e3c8d069d56623cbd2d98e811bb63c81bfa0b51d86fe2e51f322046f38b6",
            "image": "some-old-build-image"
        },
        "runImage": {
            "latestImage": "some-registry.com/old-repo/run@sha256:f845e3c8d069d56623cbd2d98e811bb63c81bfa0b51d86fe2e51f322046f38b6",
            "image": "some-old-run-image"
        }
    }
}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-new-build-image",
					"--run-image", "some-new-run-image",
					"--output", "json",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStack{
							ObjectMeta: stack.ObjectMeta,
							Spec: v1alpha1.ClusterStackSpec{
								Id: "some-new-id",
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.com/some-repo/build@" + newBuildImageId,
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.com/some-repo/run@" + newRunImageId,
								},
							},
							Status: stack.Status,
						},
					},
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Uploading to 'some-registry.com/some-repo'...
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not update the clusterstack and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-new-build-image",
					"--run-image", "some-new-run-image",
					"--dry-run",
				},
				ExpectedOutput: `Uploading to 'some-registry.com/some-repo'...
ClusterStack "some-stack" Updated (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update the clusterstack and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-registry.com/some-repo/build@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
  id: some-new-id
  runImage:
    image: some-registry.com/some-repo/run@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
status:
  buildImage:
    image: some-old-build-image
    latestImage: some-registry.com/old-repo/build@sha256:f845e3c8d069d56623cbd2d98e811bb63c81bfa0b51d86fe2e51f322046f38b6
  id: some-old-id
  runImage:
    image: some-old-run-image
    latestImage: some-registry.com/old-repo/run@sha256:f845e3c8d069d56623cbd2d98e811bb63c81bfa0b51d86fe2e51f322046f38b6
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						stack,
					},
					Args: []string{
						"some-stack",
						"--build-image", "some-new-build-image",
						"--run-image", "some-new-run-image",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Uploading to 'some-registry.com/some-repo'...
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}

func makeStackImages(t *testing.T, stackId string) (v1.Image, string, v1.Image, string) {
	buildImage, err := random.Image(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	buildImage, err = imagehelpers.SetStringLabel(buildImage, clusterstack.IdLabel, stackId)
	if err != nil {
		t.Fatal(err)
	}

	runImage, err := random.Image(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	runImage, err = imagehelpers.SetStringLabel(runImage, clusterstack.IdLabel, stackId)
	if err != nil {
		t.Fatal(err)
	}

	buildImageHash, err := buildImage.Digest()
	if err != nil {
		t.Fatal(err)
	}

	runImageHash, err := runImage.Digest()
	if err != nil {
		t.Fatal(err)
	}

	return buildImage, buildImageHash.String(), runImage, runImageHash.String()
}
