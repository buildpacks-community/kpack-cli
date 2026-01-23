// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import_test

import (
	"io/ioutil"
	"testing"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	commandsfakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	importcmds "github.com/buildpacks-community/kpack-cli/pkg/commands/import"
	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
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
		registryfakes.BuildpackImgInfo{
			Id: "my-buildpack",
			ImageInfo: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/standalone-buildpack",
				Digest: "standalone-buildpack-digest",
			},
		},
	)

	kpConfig := &corev1.ConfigMap{
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

	lifecycleImageConfig := &v1alpha2.ClusterLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "default",
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterLifecycleSpec{
			ImageSource: corev1alpha1.ImageSource{
				Image: "old/image",
			},
		},
	}

	timestampProvider := FakeTimestampProvider{timestamp: "2006-01-02T15:04:05Z"}

	clusterBuildpack := &v1alpha2.ClusterBuildpack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterBuildpackKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-buildpack",
			Annotations: map[string]string{
				importTimestampKey: timestampProvider.timestamp,
				"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterBuildpack","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"my-buildpack","creationTimestamp":null},"spec":{"image":"default-registry.io/default-repo@sha256:standalone-buildpack-digest","serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{}}`,
			},
		},
		Spec: v1alpha2.ClusterBuildpackSpec{
			ImageSource: corev1alpha1.ImageSource{
				Image: "default-registry.io/default-repo@sha256:standalone-buildpack-digest",
			},
			ServiceAccountRef: &corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
			},
		},
	}

	store := &v1alpha2.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "store-name",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"default-registry.io/default-repo@sha256:buildpack-image-digest"}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{}}`,
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha2.ClusterStoreSpec{
			ServiceAccountRef: &corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
			},
			Sources: []corev1alpha1.ImageSource{
				{Image: "default-registry.io/default-repo@sha256:buildpack-image-digest"},
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
				Image: "default-registry.io/default-repo@sha256:build-image-digest",
			},
			RunImage: v1alpha2.ClusterStackSpecImage{
				Image: "default-registry.io/default-repo@sha256:build-image-digest",
			},
			ServiceAccountRef: &corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
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
				Tag: "default-registry.io/default-repo:clusterbuilder-clusterbuilder-name",
				Stack: corev1.ObjectReference{
					Name: "stack-name",
					Kind: v1alpha2.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: "store-name",
					Kind: v1alpha2.ClusterStoreKind,
				},
				Order: []v1alpha2.BuilderOrderEntry{
					{
						Group: []v1alpha2.BuilderBuildpackRef{
							{
								BuildpackRef: corev1alpha1.BuildpackRef{
									BuildpackInfo: corev1alpha1.BuildpackInfo{
										Id: "buildpack-id",
									},
								},
							},
						},
					},
				},
			},
			ServiceAccountRef: corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
			},
		},
	}

	defaultBuilder := builder.DeepCopy()
	defaultBuilder.Name = "default"
	defaultBuilder.Spec.Tag = "default-registry.io/default-repo:clusterbuilder-default"

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
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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
				ExpectedOutput: `Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
				ExpectCreates: []runtime.Object{
					clusterBuildpack,
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectPatches: []string{
					`{"metadata":{"annotations":{"kpack.io/import-timestamp":"2006-01-02T15:04:05Z","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 7)
			require.Len(t, fakeWaiter.WaitCalls[5].ExtraChecks, 1) // ClusterBuilder has extra check
			require.Len(t, fakeWaiter.WaitCalls[6].ExtraChecks, 1) // ClusterBuilder has extra check
		})

		it("creates stores, stacks, and cbs defined in the dependency descriptor provided by stdin", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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
				ExpectedOutput: `Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
				ExpectCreates: []runtime.Object{
					clusterBuildpack,
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectPatches: []string{
					`{"metadata":{"annotations":{"kpack.io/import-timestamp":"2006-01-02T15:04:05Z","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 7)
			require.Len(t, fakeWaiter.WaitCalls[5].ExtraChecks, 1) // ClusterBuilder has extra check
			require.Len(t, fakeWaiter.WaitCalls[6].ExtraChecks, 1) // ClusterBuilder has extra check
		})

		it("creates stores, stacks, and cbs defined in the dependency descriptor for version 1", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
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
				builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
				defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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

ClusterLifecycles

some-diff

ClusterBuildpacks

some-diff

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectCreates: []runtime.Object{
						clusterBuildpack,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"2006-01-02T15:04:05Z","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
					},
				}.TestK8sAndKpack(t, cmdFunc)
				require.NoError(t, fakeConfirmationProvider.WasRequestedWithMsg("Confirm with y:"))
			})

			it("skips confirmation when the force flag is used", func() {
				builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
				defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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

ClusterLifecycles

some-diff

ClusterBuildpacks

some-diff

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectCreates: []runtime.Object{
						clusterBuildpack,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"2006-01-02T15:04:05Z","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
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

			expectedClusterBuildpack := clusterBuildpack.DeepCopy()
			expectedClusterBuildpack.Annotations[importTimestampKey] = newTimestamp

			fakeDiffer.DiffResult = ""

			it("updates the import timestamp and uses descriptive confirm message", func() {
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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
					ExpectedOutput: `Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
	Buildpackage already exists in the store
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectCreates: []runtime.Object{
						expectedClusterBuildpack,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp"}},"spec":{"buildImage":{"image":"default-registry.io/default-repo@sha256:build-image-digest"},"runImage":{"image":"default-registry.io/default-repo@sha256:build-image-digest"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"clusterbuilder-name\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"}}}`,
					},
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 7)
				require.Len(t, fakeWaiter.WaitCalls[5].ExtraChecks, 1) // ClusterBuilder has extra check
				require.Len(t, fakeWaiter.WaitCalls[6].ExtraChecks, 1) // ClusterBuilder has extra check
			})

			it("does not error when original resource annotation is nil", func() {
				lifecycleImageConfig.Annotations = nil
				store.Annotations = nil
				stack.Annotations = nil
				defaultStack.Annotations = nil
				builder.Annotations = nil
				defaultBuilder.Annotations = nil

				expectedStore.Annotations = map[string]string{importTimestampKey: newTimestamp}
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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
					ExpectedOutput: `Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
	Buildpackage already exists in the store
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectCreates: []runtime.Object{
						expectedClusterBuildpack,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"clusterbuilder-name\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"}}}`,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("the dependency descriptor has different resources", func() {
			const newTimestamp = "new-timestamp"
			timestampProvider.timestamp = newTimestamp

			expectedStore := store.DeepCopy()
			expectedStore.Annotations[importTimestampKey] = newTimestamp
			expectedStore.Spec.Sources = append(expectedStore.Spec.Sources, corev1alpha1.ImageSource{
				Image: "default-registry.io/default-repo@sha256:another-buildpack-image-digest",
			})

			expectedStack := stack.DeepCopy()
			expectedStack.Annotations[importTimestampKey] = newTimestamp
			expectedStack.Spec.Id = "another-stack-id"
			expectedStack.Spec.BuildImage.Image = "default-registry.io/default-repo@sha256:another-build-image-digest"
			expectedStack.Spec.RunImage.Image = "default-registry.io/default-repo@sha256:another-run-image-digest"

			expectedDefaultStack := defaultStack.DeepCopy()
			expectedDefaultStack.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultStack.Spec.Id = "another-stack-id"
			expectedDefaultStack.Spec.BuildImage.Image = "default-registry.io/default-repo@sha256:another-build-image-digest"
			expectedDefaultStack.Spec.RunImage.Image = "default-registry.io/default-repo@sha256:another-run-image-digest"

			expectedBuilder := builder.DeepCopy()
			expectedBuilder.Annotations[importTimestampKey] = newTimestamp
			expectedBuilder.Spec.Order = []v1alpha2.BuilderOrderEntry{
				{
					Group: []v1alpha2.BuilderBuildpackRef{
						{
							BuildpackRef: corev1alpha1.BuildpackRef{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id: "another-buildpack-id",
								},
							},
						},
					},
				},
			}

			expectedDefaultBuilder := defaultBuilder.DeepCopy()
			expectedDefaultBuilder.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultBuilder.Spec.Order = []v1alpha2.BuilderOrderEntry{
				{
					Group: []v1alpha2.BuilderBuildpackRef{
						{
							BuildpackRef: corev1alpha1.BuildpackRef{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id: "another-buildpack-id",
								},
							},
						},
					},
				},
			}

			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"another-buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
			expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"another-buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

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
					ExpectedOutput: `Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:another-lifecycle-image-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:another-buildpack-image-digest'
	Added Buildpackage
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:another-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:another-run-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:another-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:another-run-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"another-buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"}},"spec":{"order":[{"group":[{"id":"another-buildpack-id"}]}]}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:another-lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:another-lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp"}},"spec":{"sources":[{"image":"default-registry.io/default-repo@sha256:buildpack-image-digest"},{"image":"default-registry.io/default-repo@sha256:another-buildpack-image-digest"}]}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp"}},"spec":{"buildImage":{"image":"default-registry.io/default-repo@sha256:another-build-image-digest"},"id":"another-stack-id","runImage":{"image":"default-registry.io/default-repo@sha256:another-run-image-digest"}}}`,
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"new-timestamp","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"clusterbuilder-name\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"another-buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"}},"spec":{"order":[{"group":[{"id":"another-buildpack-id"}]}]}}`,
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
			ExpectedErrorOutput: "Error: did not find expected apiVersion, must be one of: [kp.kpack.io/v1alpha1 kp.kpack.io/v1alpha3 kp.kpack.io/v1]\n",
			ExpectErr:           true,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		const expectedOutput = `Importing ClusterLifecycle 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'...
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterBuilder 'clusterbuilder-name'...
Importing ClusterBuilder 'default'...
`

		builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`
		defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}`

		when("yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterLifecycle
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterLifecycle","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}'
  creationTimestamp: null
  name: default
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
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuildpack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuildpack","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"my-buildpack","creationTimestamp":null},"spec":{"image":"default-registry.io/default-repo@sha256:standalone-buildpack-digest","serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{}}'
  creationTimestamp: null
  name: my-buildpack
spec:
  image: default-registry.io/default-repo@sha256:standalone-buildpack-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"default-registry.io/default-repo@sha256:buildpack-image-digest"}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  sources:
  - image: default-registry.io/default-repo@sha256:buildpack-image-digest
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
    image: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
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
    image: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"clusterbuilder-name","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}'
  creationTimestamp: null
  name: clusterbuilder-name
spec:
  lifecycle: {}
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: default-registry.io/default-repo:clusterbuilder-clusterbuilder-name
status:
  lifecycle:
    api: {}
    apis:
      buildpack:
        deprecated: null
        supported: null
      platform:
        deprecated: null
        supported: null
    image: {}
  stack: {}
---
apiVersion: kpack.io/v1alpha2
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"default-registry.io/default-repo:clusterbuilder-default","stack":{"kind":"ClusterStack","name":"stack-name"},"lifecycle":{},"store":{"kind":"ClusterStore","name":"store-name"},"order":[{"group":[{"id":"buildpack-id"}]}],"serviceAccountRef":{"namespace":"some-namespace","name":"some-serviceaccount"}},"status":{"stack":{},"lifecycle":{"image":{},"api":{},"apis":{"buildpack":{"deprecated":null,"supported":null},"platform":{"deprecated":null,"supported":null}}}}}'
  creationTimestamp: null
  name: default
spec:
  lifecycle: {}
  order:
  - group:
    - id: buildpack-id
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  stack:
    kind: ClusterStack
    name: stack-name
  store:
    kind: ClusterStore
    name: store-name
  tag: default-registry.io/default-repo:clusterbuilder-default
status:
  lifecycle:
    api: {}
    apis:
      buildpack:
        deprecated: null
        supported: null
      platform:
        deprecated: null
        supported: null
    image: {}
  stack: {}
`

			it("can output yaml", func() {
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
						clusterBuildpack,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/import-timestamp":"2006-01-02T15:04:05Z","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("dry-run flag is used", func() {
				const expectedOutput = `Importing ClusterLifecycle 'default'... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:build-image-digest'
	Skipping 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'... (dry run)
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:build-image-digest'
	Skipping 'default-registry.io/default-repo@sha256:build-image-digest'
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

			when("dry-run-with-image-upload flag is used", func() {
				const expectedOutput = `Importing ClusterLifecycle 'default'... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:lifecycle-image-digest'
Importing ClusterBuildpack 'my-buildpack'... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:standalone-buildpack-digest'
Importing ClusterStore 'store-name'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:buildpack-image-digest'
Importing ClusterStack 'stack-name'... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
Importing ClusterStack 'default'... (dry run with image upload)
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
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

		it("can output in json format", func() {
			const resourceJSON = `[{
    "kind": "ClusterLifecycle",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "default",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"
        }
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
},{
    "kind": "ClusterBuildpack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "my-buildpack",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuildpack\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"my-buildpack\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:standalone-buildpack-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{}}"
        }
    },
    "spec": {
        "image": "default-registry.io/default-repo@sha256:standalone-buildpack-digest",
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {}
},{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "store-name",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterStore\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"store-name\",\"creationTimestamp\":null},\"spec\":{\"sources\":[{\"image\":\"default-registry.io/default-repo@sha256:buildpack-image-digest\"}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{}}"
        }
    },
    "spec": {
        "sources": [
            {
                "image": "default-registry.io/default-repo@sha256:buildpack-image-digest"
            }
        ],
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {}
},{
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
            "image": "default-registry.io/default-repo@sha256:build-image-digest"
        },
        "runImage": {
            "image": "default-registry.io/default-repo@sha256:build-image-digest"
        },
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
},{
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
            "image": "default-registry.io/default-repo@sha256:build-image-digest"
        },
        "runImage": {
            "image": "default-registry.io/default-repo@sha256:build-image-digest"
        },
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
},{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "clusterbuilder-name",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"clusterbuilder-name\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-clusterbuilder-name\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"
        }
    },
    "spec": {
        "tag": "default-registry.io/default-repo:clusterbuilder-clusterbuilder-name",
        "stack": {
            "kind": "ClusterStack",
            "name": "stack-name"
        },
        "lifecycle": {},
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
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {},
        "lifecycle": {
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
},{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "default",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"default-registry.io/default-repo:clusterbuilder-default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"stack-name\"},\"lifecycle\":{},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"store-name\"},\"order\":[{\"group\":[{\"id\":\"buildpack-id\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{},\"lifecycle\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}}"
        }
    },
    "spec": {
        "tag": "default-registry.io/default-repo:clusterbuilder-default",
        "stack": {
            "kind": "ClusterStack",
            "name": "stack-name"
        },
        "lifecycle": {},
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
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "stack": {},
        "lifecycle": {
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
}]
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
					clusterBuildpack,
					store,
					stack,
					defaultStack,
					builder,
					defaultBuilder,
				},
				ExpectPatches: []string{
					`{"metadata":{"annotations":{"kpack.io/import-timestamp":"2006-01-02T15:04:05Z","kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterLifecycle\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"image\":\"default-registry.io/default-repo@sha256:lifecycle-image-digest\",\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"image\":{},\"api\":{},\"apis\":{\"buildpack\":{\"deprecated\":null,\"supported\":null},\"platform\":{\"deprecated\":null,\"supported\":null}}}}"}},"spec":{"image":"default-registry.io/default-repo@sha256:lifecycle-image-digest","serviceAccountRef":{"name":"some-serviceaccount","namespace":"some-namespace"}}}`,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})
}

type FakeTimestampProvider struct {
	timestamp string
}

func (f FakeTimestampProvider) GetTimestamp() string {
	return f.timestamp
}
