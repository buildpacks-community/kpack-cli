// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
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
	clusterstackcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestCreateCommand(t *testing.T) {
	spec.Run(t, "TestCreateCommand", testCreateCommand)
}

func testCreateCommand(t *testing.T, when spec.G, it spec.S) {
	stackInfo := registryfakes.StackInfo{
		StackID: "stack-id",
		BuildImg: registryfakes.ImageInfo{
			Ref:    "some-registry.io/repo/some-build-image",
			Digest: "build-image-digest",
		},
		RunImg: registryfakes.ImageInfo{
			Ref:    "some-registry.io/repo/some-run-image",
			Digest: "run-image-digest",
		},
	}

	fakeRelocator := &registryfakes.Relocator{}
	fakeRegistryUtilProvider := &registryfakes.UtilProvider{
		FakeRelocator: fakeRelocator,
		FakeFetcher:   registryfakes.NewStackImagesFetcher(stackInfo),
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "canonical-registry.io/canonical-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterstackcmds.NewCreateCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	expectedStack := &v1alpha1.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "stack-name",
			Annotations: nil,
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: "stack-id",
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
			},
		},
	}

	it("creates a stack", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				"stack-name",
				"--build-image", "some-registry.io/repo/some-build-image",
				"--run-image", "some-registry.io/repo/some-run-image",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectedOutput: `Creating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
ClusterStack "stack-name" created
`,
			ExpectCreates: []runtime.Object{
				expectedStack,
			},
		}.TestK8sAndKpack(t, cmdFunc)
		require.Len(t, fakeWaiter.WaitCalls, 1)
	})

	it("fails when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				"stack-name",
				"--build-image", "some-registry.io/repo/some-build-image",
				"--run-image", "some-registry.io/repo/some-run-image",
			},
			ExpectErr: true,
			ExpectedOutput: "Creating ClusterStack...\nError: configmaps \"kp-config\" not found\n",
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
				"stack-name",
				"--build-image", "some-registry.io/repo/some-build-image",
				"--run-image", "some-registry.io/repo/some-run-image",
			},
			ExpectErr: true,
			ExpectedOutput: "Creating ClusterStack...\nError: key \"canonical.repository\" not found in configmap \"kp-config\"\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage: {}
  runImage: {}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/some-build-image",
					"--run-image", "some-registry.io/repo/some-run-image",
					"--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Creating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
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
        "name": "stack-name",
        "creationTimestamp": null
    },
    "spec": {
        "id": "stack-id",
        "buildImage": {
            "image": "canonical-registry.io/canonical-repo/build@sha256:build-image-digest"
        },
        "runImage": {
            "image": "canonical-registry.io/canonical-repo/run@sha256:run-image-digest"
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
					"stack-name",
					"--build-image", "some-registry.io/repo/some-build-image",
					"--run-image", "some-registry.io/repo/some-run-image",
					"--output", "json",
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Creating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		fakeRelocator.SetSkip(true)

		it("does not create a clusterstack and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/some-build-image",
					"--run-image", "some-registry.io/repo/some-run-image",
					"--dry-run",
				},
				ExpectedOutput: `Creating ClusterStack... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
ClusterStack "stack-name" created (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 0)
		})

		when("output flag is used", func() {
			it("does not create a clusterstack and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage: {}
  runImage: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/some-build-image",
						"--run-image", "some-registry.io/repo/some-run-image",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterStack... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run-with-image-upload flag is used", func() {
		it("does not create a clusterstack and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/some-build-image",
					"--run-image", "some-registry.io/repo/some-run-image",
					"--dry-run-with-image-upload",
				},
				ExpectedOutput: `Creating ClusterStack... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
ClusterStack "stack-name" created (dry run with image upload)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not create a clusterstack and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage: {}
  runImage: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/some-build-image",
						"--run-image", "some-registry.io/repo/some-run-image",
						"--dry-run-with-image-upload",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterStack... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}
