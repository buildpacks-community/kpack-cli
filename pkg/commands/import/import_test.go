// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import_test

import (
	"io/ioutil"
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	importcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/import"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestImportCommand(t *testing.T) {
	spec.Run(t, "TestImportCommand", testImportCommand)
}

func testImportCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		lifecycleImageKey  = "image"
		importTimestampKey = "kpack.io/import-timestamp"
	)

	fakeFetcher := &registryfakes.Fetcher{}
	fakeRegistryUtilProvider := &registryfakes.UtilProvider{
		FakeFetcher: fakeFetcher,
	}

	fakeFetcher.AddLifecycleImages(
		registryfakes.LifecycleInfo{
			Metadata: "value-not-validated-by-cli",
			ImageInfo: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/lifecycle-image",
				Digest: "lifecycle-image-digest",
			},
		},
		registryfakes.LifecycleInfo{
			Metadata: "value-not-validated-by-cli",
			ImageInfo: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/another-lifecycle-image",
				Digest: "another-lifecycle-image-digest",
			},
		},
	)

	fakeFetcher.AddStackImages(
		registryfakes.StackInfo{
			StackID: "stack-id",
			BuildImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/build-image",
				Digest: "build-image-digest",
			},
			RunImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/run-image",
				Digest: "build-image-digest",
			},
		},
		registryfakes.StackInfo{
			StackID: "another-stack-id",
			BuildImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/another-build-image",
				Digest: "another-build-image-digest",
			},
			RunImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/another-run-image",
				Digest: "another-run-image-digest",
			},
		},
	)

	fakeFetcher.AddBuildpackImages(
		registryfakes.BuildpackImgInfo{
			Id: "buildpack-id",
			ImageInfo: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/buildpack-image",
				Digest: "buildpack-image-digest",
			},
		},
		registryfakes.BuildpackImgInfo{
			Id: "another-buildpack-id",
			ImageInfo: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/another-buildpack-image",
				Digest: "another-buildpack-image-digest",
			},
		},
	)

	kpConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "canonical-registry.io/canonical-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	lifecycleImageConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "lifecycle-image",
			Namespace:   "kpack",
			Annotations: map[string]string{},
		},
		Data: map[string]string{},
	}

	timestampProvider := FakeTimestampProvider{timestamp: "2006-01-02T15:04:05Z"}

	expectedLifecycleImageConfig := lifecycleImageConfig.DeepCopy()
	expectedLifecycleImageConfig.Annotations[importTimestampKey] = timestampProvider.timestamp
	expectedLifecycleImageConfig.Data["image"] = "canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest"

	store := &v1alpha2.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "store-name",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest"}]},"status":{}}`,
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha2.ClusterStoreSpec{
			Sources: []v1alpha2.StoreImage{
				{Image: "canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest"},
			},
		},
	}

	stack := &v1alpha2.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "stack-name",
			Annotations: map[string]string{
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha2.ClusterStackSpec{
			Id: "stack-id",
			BuildImage: v1alpha2.ClusterStackSpecImage{
				Image: "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
			},
			RunImage: v1alpha2.ClusterStackSpecImage{
				Image: "canonical-registry.io/canonical-repo/run@sha256:build-image-digest",
			},
		},
	}

	defaultStack := stack.DeepCopy()
	defaultStack.Name = "default"

	builder := &v1alpha2.ClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterBuilderKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusterbuilder-name",
			Annotations: map[string]string{
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha2.ClusterBuilderSpec{
			BuilderSpec: v1alpha2.BuilderSpec{
				Tag: "canonical-registry.io/canonical-repo/clusterbuilder-name",
				Stack: corev1.ObjectReference{
					Name: "stack-name",
					Kind: v1alpha2.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: "store-name",
					Kind: v1alpha2.ClusterStoreKind,
				},
				Order: []v1alpha2.OrderEntry{
					{
						Group: []v1alpha2.BuildpackRef{
							{
								BuildpackInfo: v1alpha2.BuildpackInfo{
									Id: "buildpack-id",
								},
							},
						},
					},
				},
			},
			ServiceAccountRef: corev1.ObjectReference{
				Namespace: "kpack",
				Name:      "some-serviceaccount",
			},
		},
	}

	defaultBuilder := builder.DeepCopy()
	defaultBuilder.Name = "default"
	defaultBuilder.Spec.Tag = "canonical-registry.io/canonical-repo/default"

	var fakeConfirmationProvider *commandsfakes.FakeConfirmationProvider
	fakeDiffer := &commandsfakes.FakeDiffer{DiffResult: "some-diff"}

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return importcmds.NewImportCommand(
			fakeDiffer,
			clientSetProvider,
			fakeRegistryUtilProvider,
			timestampProvider,
			fakeConfirmationProvider,
			func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			},
		)
	}

	it.Before(func() {
		fakeConfirmationProvider = commandsfakes.NewFakeConfirmationProvider(true, nil)
	})

	when("there are no stores, stacks, or cbs", func() {
		it("creates stores, stacks, and cbs defined in the dependency descriptor", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectedOutput: `Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
				ExpectCreates: []runtime.Object{
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: expectedLifecycleImageConfig,
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 5)
			require.Len(t, fakeWaiter.WaitCalls[3].ExtraChecks, 1) // ClusterBuilder has extra check
			require.Len(t, fakeWaiter.WaitCalls[4].ExtraChecks, 1) // ClusterBuilder has extra check
		})

		it("creates stores, stacks, and cbs defined in the dependency descriptor provided by stdin", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			descriptor, err := ioutil.ReadFile("./testdata/deps.yaml")
			require.NoError(t, err)

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "-",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				StdIn: string(descriptor),
				ExpectedOutput: `Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
				ExpectCreates: []runtime.Object{
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: expectedLifecycleImageConfig,
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 5)
			require.Len(t, fakeWaiter.WaitCalls[3].ExtraChecks, 1) // ClusterBuilder has extra check
			require.Len(t, fakeWaiter.WaitCalls[4].ExtraChecks, 1) // ClusterBuilder has extra check
		})

		it("creates stores, stacks, and cbs defined in the dependency descriptor for version 1", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "./testdata/v1-deps.yaml",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectedOutput: `Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
				ExpectCreates: []runtime.Object{
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("the show changes flag is used", func() {
			it("shows a summary of changes for each resource", func() {
				builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
				defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"--show-changes",
					},
					ExpectedOutput: `Changes

Lifecycle

some-diff

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectCreates: []runtime.Object{
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedLifecycleImageConfig,
						},
					},
				}.TestK8sAndKpack(t, cmdFunc)
				require.NoError(t, fakeConfirmationProvider.WasRequestedWithMsg("Confirm with y:"))
			})

			it("skips confirmation when the force flag is used", func() {
				builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
				defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"--show-changes",
						"--force",
					},
					ExpectedOutput: `Changes

Lifecycle

some-diff

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectCreates: []runtime.Object{
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedLifecycleImageConfig,
						},
					},
				}.TestK8sAndKpack(t, cmdFunc)
				require.Equal(t, false, fakeConfirmationProvider.WasRequested())
			})
		})
	})

	when("there are existing stores, stacks, or cbs", func() {
		when("the dependency descriptor and the cluster have the exact same objs", func() {
			const newTimestamp = "new-timestamp"
			timestampProvider.timestamp = newTimestamp

			expectedLifecycleImageConfig.Annotations[importTimestampKey] = newTimestamp

			store.Generation = 12
			expectedStore := store.DeepCopy()
			expectedStore.Annotations[importTimestampKey] = newTimestamp

			stack.Generation = 13
			expectedStack := stack.DeepCopy()
			expectedStack.Annotations[importTimestampKey] = newTimestamp

			expectedDefaultStack := defaultStack.DeepCopy()
			expectedDefaultStack.Annotations[importTimestampKey] = newTimestamp

			expectedBuilder := builder.DeepCopy()
			expectedBuilder.Annotations[importTimestampKey] = newTimestamp

			expectedDefaultBuilder := defaultBuilder.DeepCopy()
			expectedDefaultBuilder.Annotations[importTimestampKey] = newTimestamp

			fakeDiffer.DiffResult = ""

			it("updates the import timestamp and uses descriptive confirm message", func() {
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				stack.Spec.BuildImage.Image = "some-uploaded-build-image@build-image-digest"
				stack.Spec.RunImage.Image = "some-uploaded-run-image@build-image-digest"

				defaultStack.Spec.BuildImage.Image = "some-uploaded-build-image@build-image-digest"
				defaultStack.Spec.RunImage.Image = "some-uploaded-run-image@build-image-digest"

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
					},
					ExpectedOutput: `Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
	Buildpackage already exists in the store
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{Object: expectedLifecycleImageConfig},
						{Object: expectedStore},
						{Object: expectedStack},
						{Object: expectedDefaultStack},
						{Object: expectedBuilder},
						{Object: expectedDefaultBuilder},
					},
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 5)
				require.Len(t, fakeWaiter.WaitCalls[3].ExtraChecks, 1) // ClusterBuilder has extra check
				require.Len(t, fakeWaiter.WaitCalls[4].ExtraChecks, 1) // ClusterBuilder has extra check
			})

			it("does not error when original resource annotation is nil", func() {
				lifecycleImageConfig.Annotations = nil
				store.Annotations = nil
				stack.Annotations = nil
				defaultStack.Annotations = nil
				builder.Annotations = nil
				defaultBuilder.Annotations = nil

				expectedLifecycleImageConfig.Annotations = map[string]string{importTimestampKey: newTimestamp}
				expectedStore.Annotations = map[string]string{importTimestampKey: newTimestamp}
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
					},
					ExpectedOutput: `Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
	Buildpackage already exists in the store
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{Object: expectedLifecycleImageConfig},
						{Object: expectedStore},
						{Object: expectedStack},
						{Object: expectedDefaultStack},
						{Object: expectedBuilder},
						{Object: expectedDefaultBuilder},
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("the dependency descriptor has different resources", func() {
			const newTimestamp = "new-timestamp"
			timestampProvider.timestamp = newTimestamp

			expectedLifecycleImageConfig.Annotations[importTimestampKey] = newTimestamp
			expectedLifecycleImageConfig.Data[lifecycleImageKey] = "canonical-registry.io/canonical-repo/lifecycle@sha256:another-lifecycle-image-digest"

			expectedStore := store.DeepCopy()
			expectedStore.Annotations[importTimestampKey] = newTimestamp
			expectedStore.Spec.Sources = append(expectedStore.Spec.Sources, v1alpha2.StoreImage{
				Image: "canonical-registry.io/canonical-repo/another-buildpack-id@sha256:another-buildpack-image-digest",
			})

			expectedStack := stack.DeepCopy()
			expectedStack.Annotations[importTimestampKey] = newTimestamp
			expectedStack.Spec.Id = "another-stack-id"
			expectedStack.Spec.BuildImage.Image = "canonical-registry.io/canonical-repo/build@sha256:another-build-image-digest"
			expectedStack.Spec.RunImage.Image = "canonical-registry.io/canonical-repo/run@sha256:another-run-image-digest"

			expectedDefaultStack := defaultStack.DeepCopy()
			expectedDefaultStack.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultStack.Spec.Id = "another-stack-id"
			expectedDefaultStack.Spec.BuildImage.Image = "canonical-registry.io/canonical-repo/build@sha256:another-build-image-digest"
			expectedDefaultStack.Spec.RunImage.Image = "canonical-registry.io/canonical-repo/run@sha256:another-run-image-digest"

			expectedBuilder := builder.DeepCopy()
			expectedBuilder.Annotations[importTimestampKey] = newTimestamp
			expectedBuilder.Spec.Order = []v1alpha2.OrderEntry{
				{
					Group: []v1alpha2.BuildpackRef{
						{
							BuildpackInfo: v1alpha2.BuildpackInfo{
								Id: "another-buildpack-id",
							},
						},
					},
				},
			}

			expectedDefaultBuilder := defaultBuilder.DeepCopy()
			expectedDefaultBuilder.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultBuilder.Spec.Order = []v1alpha2.OrderEntry{
				{
					Group: []v1alpha2.BuildpackRef{
						{
							BuildpackInfo: v1alpha2.BuildpackInfo{
								Id: "another-buildpack-id",
							},
						},
					},
				},
			}

			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"another-buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"another-buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			it("creates stores, stacks, and cbs defined in the dependency descriptor and updates the timestamp", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/updated-deps.yaml",
					},
					ExpectedOutput: `Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:another-lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/another-buildpack-id@sha256:another-buildpack-image-digest'
	Added Buildpackage
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:another-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:another-run-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:another-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:another-run-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{Object: expectedLifecycleImageConfig},
						{Object: expectedStore},
						{Object: expectedStack},
						{Object: expectedDefaultStack},
						{Object: expectedBuilder},
						{Object: expectedDefaultBuilder},
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	it("errors when the descriptor apiVersion is unexpected", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{kpConfig},
			Args: []string{
				"-f", "./testdata/invalid-deps.yaml",
			},
			ExpectedErrorOutput: "Error: did not find expected apiVersion, must be one of: [kp.kpack.io/v1alpha2 kp.kpack.io/v1alpha3]\n",
			ExpectErr:           true,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		const expectedOutput = `Importing Lifecycle...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
`

		builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
		defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: v1
data:
  image: canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest
kind: ConfigMap
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: lifecycle-image
  namespace: kpack
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest"}]},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest
status: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:build-image-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: default
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:build-image-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: clusterbuilder-name
spec:
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: canonical-registry.io/canonical-repo/clusterbuilder-name
status:
  stack: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: default
spec:
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: canonical-registry.io/canonical-repo/default
status:
  stack: {}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--output", "yaml",
				},
				ExpectedOutput:      resourceYAML,
				ExpectedErrorOutput: expectedOutput,
				ExpectCreates: []runtime.Object{
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: expectedLifecycleImageConfig,
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ConfigMap",
    "apiVersion": "v1",
    "metadata": {
        "name": "lifecycle-image",
        "namespace": "kpack",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z"
        }
    },
    "data": {
        "image": "canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest"
    }
}
{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "store-name",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterStore\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"store-name\",\"creationTimestamp\":null},\"spec\":{\"sources\":[{\"image\":\"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest\"}]},\"status\":{}}"
        }
    },
    "spec": {
        "sources": [
            {
                "image": "canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest"
            }
        ]
    },
    "status": {}
}
{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "stack-name",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z"
        }
    },
    "spec": {
        "id": "stack-id",
        "buildImage": {
            "image": "canonical-registry.io/canonical-repo/build@sha256:build-image-digest"
        },
        "runImage": {
            "image": "canonical-registry.io/canonical-repo/run@sha256:build-image-digest"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "default",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z"
        }
    },
    "spec": {
        "id": "stack-id",
        "buildImage": {
            "image": "canonical-registry.io/canonical-repo/build@sha256:build-image-digest"
        },
        "runImage": {
            "image": "canonical-registry.io/canonical-repo/run@sha256:build-image-digest"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "clusterbuilder-name",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"clusterbuilder-name\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"canonical-registry.io/canonical-repo/clusterbuilder-name\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "canonical-registry.io/canonical-repo/clusterbuilder-name",
        "stack": {
            "kind": "ClusterStack",
            "name": "stack-name"
        },
        "store": {
            "kind": "ClusterStore",
            "name": "store-name"
        },
        "order": [
            {
                "group": [
                    {
                        "id": "buildpack-id"
                    }
                ]
            }
        ],
        "serviceAccountRef": {
            "namespace": "kpack",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {}
    }
}
{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "default",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"canonical-registry.io/canonical-repo/default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "canonical-registry.io/canonical-repo/default",
        "stack": {
            "kind": "ClusterStack",
            "name": "stack-name"
        },
        "store": {
            "kind": "ClusterStore",
            "name": "store-name"
        },
        "order": [
            {
                "group": [
                    {
                        "id": "buildpack-id"
                    }
                ]
            }
        ],
        "serviceAccountRef": {
            "namespace": "kpack",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {}
    }
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--output", "json",
				},
				ExpectedOutput:      resourceJSON,
				ExpectedErrorOutput: expectedOutput,
				ExpectCreates: []runtime.Object{
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: expectedLifecycleImageConfig,
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
		defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		it("does not create any resources and prints result with dry run indicated", func() {
			const expectedOutput = `Importing Lifecycle... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'... (dry run)
Importing ClusterBuilder 'default'... (dry run)
Imported resources (dry run)
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--dry-run",
				},
				ExpectedOutput: expectedOutput,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 0)
		})

		when("output flag is used", func() {
			const resourceYAML = `apiVersion: v1
data:
  image: canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest
kind: ConfigMap
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: lifecycle-image
  namespace: kpack
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest"}]},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest
status: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:build-image-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: default
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:build-image-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: clusterbuilder-name
spec:
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: canonical-registry.io/canonical-repo/clusterbuilder-name
status:
  stack: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: default
spec:
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: canonical-registry.io/canonical-repo/default
status:
  stack: {}
`

			const expectedOutput = `Importing Lifecycle... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'... (dry run)
Importing ClusterBuilder 'default'... (dry run)
`

			it("does not create a Builder and prints the resource output", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput:      resourceYAML,
					ExpectedErrorOutput: expectedOutput,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run-with-image-upload flag is used", func() {
		builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
		defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		it("does not create any resources and prints result with dry run indicated", func() {
			const expectedOutput = `Importing Lifecycle... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'... (dry run with image upload)
Importing ClusterBuilder 'default'... (dry run with image upload)
Imported resources (dry run with image upload)
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--dry-run-with-image-upload",
				},
				ExpectedOutput: expectedOutput,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			const resourceYAML = `apiVersion: v1
data:
  image: canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest
kind: ConfigMap
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: lifecycle-image
  namespace: kpack
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest"}]},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest
status: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:build-image-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: default
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:build-image-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: clusterbuilder-name
spec:
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: canonical-registry.io/canonical-repo/clusterbuilder-name
status:
  stack: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"canonical-registry.io/canonical-repo/default","stack":{"kind":"ClusterStack","name":"stack-name"},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: default
spec:
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: canonical-registry.io/canonical-repo/default
status:
  stack: {}
`

			const expectedOutput = `Importing Lifecycle... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Importing ClusterStore 'store-name'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterStack 'default'... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'... (dry run with image upload)
Importing ClusterBuilder 'default'... (dry run with image upload)
`

			it("does not create a Builder and prints the resource output", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"--dry-run-with-image-upload",
						"--output", "yaml",
					},
					ExpectedOutput:      resourceYAML,
					ExpectedErrorOutput: expectedOutput,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}

type FakeTimestampProvider struct {
	timestamp string
}

func (f FakeTimestampProvider) GetTimestamp() string {
	return f.timestamp
}
