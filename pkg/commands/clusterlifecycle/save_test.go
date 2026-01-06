// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterlifecycle"
	commandsfakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestSaveCommand(t *testing.T) {
	spec.Run(t, "TestSaveCommandCreate", testSaveCreateCommand)
	spec.Run(t, "TestSaveCommandPatch", testSavePatchCommand)
}

func testSaveCreateCommand(t *testing.T, when spec.G, it spec.S) {
	lifecycleImageInfo := registryfakes.LifecycleInfo{
		Version: "0.17.0",
		Apis:    `{"buildpack":{"deprecated":[],"supported":["0.2","0.3","0.4","0.5","0.6","0.7","0.8","0.9","0.10"]},"platform":{"deprecated":[],"supported":["0.3","0.4","0.5","0.6","0.7","0.8","0.9","0.10","0.11","0.12"]}}`,
		ImageInfo: registryfakes.ImageInfo{
			Ref:    "some-registry.io/repo/lifecycle",
			Digest: "lifecycle-image-digest",
		},
	}

	fakeRegistryUtilProvider := &registryfakes.UtilProvider{
		FakeFetcher: registryfakes.NewLifecycleImageFetcher(lifecycleImageInfo),
	}

	config := &corev1.ConfigMap{
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

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterlifecycle.NewSaveCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	expectedLifecycle := &v1alpha2.ClusterLifecycle{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterLifecycleKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-lifecycle",
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterLifecycleSpec{
			ImageSource: v1alpha1.ImageSource{
				Image: "default-registry.io/default-repo@sha256:lifecycle-image-digest",
			},
			ServiceAccountRef: &corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
			},
		},
	}

	it("creates a clusterlifecycle when it does not exist", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				config,
			},
			Args: []string{
				"my-lifecycle",
				"--image", "some-registry.io/repo/lifecycle",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectedOutput: `Creating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" created
`,
			ExpectCreates: []runtime.Object{
				expectedLifecycle,
			},
		}.TestK8sAndKpack(t, cmdFunc)
		require.Len(t, fakeWaiter.WaitCalls, 1)
	})

	it("fails when default.repository key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{},
		}

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				badConfig,
			},
			Args: []string{
				"my-lifecycle",
				"--image", "some-registry.io/repo/lifecycle",
			},
			ExpectErr:           true,
			ExpectedOutput:      "Creating ClusterLifecycle...\n",
			ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  creationTimestamp: null
  name: my-lifecycle
spec:
  image: default-registry.io/default-repo@sha256:lifecycle-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  api: {}
  apis:
    buildpack:
      deprecated: null
      supported: null
    platform:
      deprecated: null
      supported: null
  image: {}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/lifecycle",
					"--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Creating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
`,
				ExpectCreates: []runtime.Object{
					expectedLifecycle,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ClusterLifecycle",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "my-lifecycle",
        "creationTimestamp": null
    },
    "spec": {
        "image": "default-registry.io/default-repo@sha256:lifecycle-image-digest",
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "image": {},
        "api": {},
        "apis": {
            "buildpack": {
                "deprecated": null,
                "supported": null
            },
            "platform": {
                "deprecated": null,
                "supported": null
            }
        }
    }
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/lifecycle",
					"--output", "json",
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Creating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
`,
				ExpectCreates: []runtime.Object{
					expectedLifecycle,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not create a clusterlifecycle and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/lifecycle",
					"--dry-run",
				},
				ExpectedOutput: `Creating ClusterLifecycle... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" created (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 0)
		})

		when("output flag is used", func() {
			it("does not create a clusterlifecycle and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  creationTimestamp: null
  name: my-lifecycle
spec:
  image: default-registry.io/default-repo@sha256:lifecycle-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  api: {}
  apis:
    buildpack:
      deprecated: null
      supported: null
    platform:
      deprecated: null
      supported: null
  image: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						"my-lifecycle",
						"--image", "some-registry.io/repo/lifecycle",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterLifecycle... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run-with-image-upload flag is used", func() {
		it("does not create a clusterlifecycle and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/lifecycle",
					"--dry-run-with-image-upload",
				},
				ExpectedOutput: `Creating ClusterLifecycle... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" created (dry run with image upload)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not create a clusterlifecycle and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  creationTimestamp: null
  name: my-lifecycle
spec:
  image: default-registry.io/default-repo@sha256:lifecycle-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  api: {}
  apis:
    buildpack:
      deprecated: null
      supported: null
    platform:
      deprecated: null
      supported: null
  image: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						"my-lifecycle",
						"--image", "some-registry.io/repo/lifecycle",
						"--dry-run-with-image-upload",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterLifecycle... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}

func testSavePatchCommand(t *testing.T, when spec.G, it spec.S) {
	newLifecycleImageInfo := registryfakes.LifecycleInfo{
		Version: "0.18.0",
		Apis:    `{"buildpack":{"deprecated":[],"supported":["0.2","0.3","0.4","0.5","0.6","0.7","0.8","0.9","0.10"]},"platform":{"deprecated":[],"supported":["0.3","0.4","0.5","0.6","0.7","0.8","0.9","0.10","0.11","0.12"]}}`,
		ImageInfo: registryfakes.ImageInfo{
			Ref:    "some-registry.io/repo/new-lifecycle",
			Digest: "new-lifecycle-image-digest",
		},
	}

	fakeFetcher := registryfakes.NewLifecycleImageFetcher(newLifecycleImageInfo)

	fakeRegistryUtilProvider := &registryfakes.UtilProvider{
		FakeFetcher: fakeFetcher,
	}

	existingLifecycle := &v1alpha2.ClusterLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-lifecycle",
		},
		Spec: v1alpha2.ClusterLifecycleSpec{
			ImageSource: v1alpha1.ImageSource{
				Image: "default-registry.io/default-repo@sha256:old-lifecycle-image-digest",
			},
			ServiceAccountRef: &corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
			},
		},
	}

	config := &corev1.ConfigMap{
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

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterlifecycle.NewSaveCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	it("patches an existing clusterlifecycle", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				config,
				existingLifecycle,
			},
			Args: []string{
				"my-lifecycle",
				"--image", "some-registry.io/repo/new-lifecycle",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectPatches: []string{
				`{"spec":{"image":"default-registry.io/default-repo@sha256:new-lifecycle-image-digest"}}`,
			},
			ExpectedOutput: `Updating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" updated
`,
		}.TestK8sAndKpack(t, cmdFunc)
		require.Len(t, fakeWaiter.WaitCalls, 1)
	})

	it("returns error when default.repository key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{},
		}

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				badConfig,
				existingLifecycle,
			},
			Args: []string{
				"my-lifecycle",
				"--image", "some-registry.io/repo/new-lifecycle",
			},
			ExpectErr:           true,
			ExpectedOutput:      "Updating ClusterLifecycle...\n",
			ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  creationTimestamp: null
  name: my-lifecycle
spec:
  image: default-registry.io/default-repo@sha256:new-lifecycle-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  api: {}
  apis:
    buildpack:
      deprecated: null
      supported: null
    platform:
      deprecated: null
      supported: null
  image: {}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					existingLifecycle,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/new-lifecycle",
					"--output", "yaml",
				},
				ExpectPatches: []string{
					`{"spec":{"image":"default-registry.io/default-repo@sha256:new-lifecycle-image-digest"}}`,
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Updating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ClusterLifecycle",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "my-lifecycle",
        "creationTimestamp": null
    },
    "spec": {
        "image": "default-registry.io/default-repo@sha256:new-lifecycle-image-digest",
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "image": {},
        "api": {},
        "apis": {
            "buildpack": {
                "deprecated": null,
                "supported": null
            },
            "platform": {
                "deprecated": null,
                "supported": null
            }
        }
    }
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					existingLifecycle,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/new-lifecycle",
					"--output", "json",
				},
				ExpectPatches: []string{
					`{"spec":{"image":"default-registry.io/default-repo@sha256:new-lifecycle-image-digest"}}`,
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Updating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("there are no changes in the update", func() {
			it("reports no change", func() {
				sameLifecycleImageInfo := registryfakes.LifecycleInfo{
					Version: "0.17.0",
					Apis:    `{"buildpack":{"deprecated":[],"supported":["0.2","0.3","0.4","0.5","0.6","0.7","0.8","0.9","0.10"]},"platform":{"deprecated":[],"supported":["0.3","0.4","0.5","0.6","0.7","0.8","0.9","0.10","0.11","0.12"]}}`,
					ImageInfo: registryfakes.ImageInfo{
						Ref:    "some-registry.io/repo/same-lifecycle",
						Digest: "old-lifecycle-image-digest",
					},
				}
				fakeFetcher.AddLifecycleImages(sameLifecycleImageInfo)

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingLifecycle,
					},
					Args: []string{
						"my-lifecycle",
						"--image", "some-registry.io/repo/same-lifecycle",
					},
					ExpectErr: false,
					ExpectedOutput: `Updating ClusterLifecycle...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:old-lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" updated (no change)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run flag is used", func() {
		it("does not update the clusterlifecycle and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					existingLifecycle,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/new-lifecycle",
					"--dry-run",
				},
				ExpectedOutput: `Updating ClusterLifecycle... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" updated (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 0)
		})

		when("output flag is used", func() {
			it("does not update the clusterlifecycle and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  creationTimestamp: null
  name: my-lifecycle
spec:
  image: default-registry.io/default-repo@sha256:new-lifecycle-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  api: {}
  apis:
    buildpack:
      deprecated: null
      supported: null
    platform:
      deprecated: null
      supported: null
  image: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingLifecycle,
					},
					Args: []string{
						"my-lifecycle",
						"--image", "some-registry.io/repo/new-lifecycle",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating ClusterLifecycle... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run-with-image-upload flag is used", func() {
		it("does not update the clusterlifecycle and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					existingLifecycle,
				},
				Args: []string{
					"my-lifecycle",
					"--image", "some-registry.io/repo/new-lifecycle",
					"--dry-run-with-image-upload",
				},
				ExpectedOutput: `Updating ClusterLifecycle... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
ClusterLifecycle "my-lifecycle" updated (dry run with image upload)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update the clusterlifecycle and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  creationTimestamp: null
  name: my-lifecycle
spec:
  image: default-registry.io/default-repo@sha256:new-lifecycle-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  api: {}
  apis:
    buildpack:
      deprecated: null
      supported: null
    platform:
      deprecated: null
      supported: null
  image: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingLifecycle,
					},
					Args: []string{
						"my-lifecycle",
						"--image", "some-registry.io/repo/new-lifecycle",
						"--dry-run-with-image-upload",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating ClusterLifecycle... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:new-lifecycle-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}
