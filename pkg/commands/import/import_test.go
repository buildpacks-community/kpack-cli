package _import_test

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
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	clusterstackfakes "github.com/pivotal/build-service-cli/pkg/clusterstack/fakes"
	clusterstorefakes "github.com/pivotal/build-service-cli/pkg/clusterstore/fakes"
	commandsfakes "github.com/pivotal/build-service-cli/pkg/commands/fakes"
	importcmds "github.com/pivotal/build-service-cli/pkg/commands/import"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImportCommand(t *testing.T) {
	spec.Run(t, "TestImportCommand", testImportCommand)
}

func testImportCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		importTimestampKey = "kpack.io/import-timestamp"
	)
	fakeBuildpackageUploader := &clusterstorefakes.FakeBuildpackageUploader{
		"some-registry.io/some-project/store-image":   "new-registry.io/new-project/store-image@sha256:123abc",
		"some-registry.io/some-project/store-image-2": "new-registry.io/new-project/store-image-2@sha256:456def",
	}

	fakeStackUploader := &clusterstackfakes.FakeStackUploader{
		Images: map[string]string{
			"some-registry.io/some-project/build-image":   "some-uploaded-build-image@some-digest",
			"some-registry.io/some-project/build-image-2": "some-uploaded-build-image-2@some-digest",
			"some-registry.io/some-project/run-image":     "some-uploaded-run-image@some-digest",
			"some-registry.io/some-project/run-image-2":   "some-uploaded-run-image-2@some-digest",
		},
		StackID: "some-stack-id",
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "new-registry.io/new-project",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	timestampProvider := FakeTimestampProvider{timestamp: "2006-01-02T15:04:05Z"}

	store := &v1alpha1.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-store",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-store","creationTimestamp":null},"spec":{"sources":[{"image":"new-registry.io/new-project/store-image@sha256:123abc"}]},"status":{}}`,
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha1.ClusterStoreSpec{
			Sources: []v1alpha1.StoreImage{
				{Image: "new-registry.io/new-project/store-image@sha256:123abc"},
			},
		},
	}

	stack := &v1alpha1.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-stack",
			Annotations: map[string]string{
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: "some-stack-id",
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: "some-uploaded-build-image@some-digest",
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: "some-uploaded-run-image@some-digest",
			},
		},
	}

	defaultStack := stack.DeepCopy()
	defaultStack.Name = "default"

	builder := &v1alpha1.ClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterBuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-cb",
			Annotations: map[string]string{
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha1.ClusterBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: "new-registry.io/new-project/some-cb",
				Stack: corev1.ObjectReference{
					Name: "some-stack",
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: "some-store",
					Kind: v1alpha1.ClusterStoreKind,
				},
				Order: []v1alpha1.OrderEntry{
					{
						Group: []v1alpha1.BuildpackRef{
							{
								BuildpackInfo: v1alpha1.BuildpackInfo{
									Id: "buildpack-1",
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
	defaultBuilder.Spec.Tag = "new-registry.io/new-project/default"

	var fakeConfirmationProvider *commandsfakes.FakeConfirmationProvider
	fakeDiffer := &commandsfakes.FakeDiffer{DiffResult: "some-diff"}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return importcmds.NewImportCommand(
			clientSetProvider,
			fakeBuildpackageUploader,
			fakeStackUploader,
			fakeDiffer,
			timestampProvider,
			fakeConfirmationProvider)
	}

	it.Before(func() {
		fakeConfirmationProvider = commandsfakes.NewFakeConfirmationProvider(true, nil)
	})

	when("there are no stores, stacks, or cbs", func() {
		it("creates stores, stacks, and cbs defined in the dependency descriptor", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectedOutput: `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'...
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
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
			require.NoError(t, fakeConfirmationProvider.WasRequestedWithMsg("Confirm with y:"))
		})

		it("creates stores, stacks, and cbs defined in the dependency descriptor for version 1", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"-f", "./testdata/v1-deps.yaml",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectedOutput: `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'...
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
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

		it("skips confirmation when the force flag is used", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
					"--force",
				},
				ExpectedOutput: `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'...
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
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
			require.Equal(t, false, fakeConfirmationProvider.WasRequested())
		})
	})

	when("there are existing stores, stacks, or cbs", func() {
		when("the dependency descriptor and the cluster have the exact same objs", func() {
			const newTimestamp = "new-timestamp"
			timestampProvider.timestamp = newTimestamp

			expectedStore := store.DeepCopy()
			expectedStore.Annotations[importTimestampKey] = newTimestamp

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
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				stack.Spec.BuildImage.Image = "some-uploaded-build-image@some-digest"
				stack.Spec.RunImage.Image = "some-uploaded-run-image@some-digest"

				defaultStack.Spec.BuildImage.Image = "some-uploaded-build-image@some-digest"
				defaultStack.Spec.RunImage.Image = "some-uploaded-run-image@some-digest"

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
					},
					ExpectedOutput: `Changes

ClusterStores

No Changes

ClusterStacks

No Changes

ClusterBuilders

No Changes


Importing ClusterStore 'some-store'...
	Buildpackage already exists in the store
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedStore,
						},
						{
							Object: expectedStack,
						},
						{
							Object: expectedDefaultStack,
						},
						{
							Object: expectedBuilder,
						},
						{
							Object: expectedDefaultBuilder,
						},
					},
				}.TestK8sAndKpack(t, cmdFunc)
				require.NoError(t, fakeConfirmationProvider.WasRequestedWithMsg("Re-upload images with y:"))
			})

			it("does not error when original resource annotation is nil", func() {
				store.Annotations = nil
				stack.Annotations = nil
				defaultStack.Annotations = nil
				builder.Annotations = nil
				defaultBuilder.Annotations = nil

				expectedStore.Annotations = map[string]string{importTimestampKey: newTimestamp}
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
					},
					ExpectedOutput: `Changes

ClusterStores

No Changes

ClusterStacks

No Changes

ClusterBuilders

No Changes


Importing ClusterStore 'some-store'...
	Buildpackage already exists in the store
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedStore,
						},
						{
							Object: expectedStack,
						},
						{
							Object: expectedDefaultStack,
						},
						{
							Object: expectedBuilder,
						},
						{
							Object: expectedDefaultBuilder,
						},
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("the dependency descriptor has different resources", func() {
			const newTimestamp = "new-timestamp"
			timestampProvider.timestamp = newTimestamp

			fakeStackUploader.StackID = "some-other-stack-id"

			expectedStore := store.DeepCopy()
			expectedStore.Annotations[importTimestampKey] = newTimestamp
			expectedStore.Spec.Sources = append(expectedStore.Spec.Sources, v1alpha1.StoreImage{
				Image: "new-registry.io/new-project/store-image-2@sha256:456def",
			})

			expectedStack := stack.DeepCopy()
			expectedStack.Annotations[importTimestampKey] = newTimestamp
			expectedStack.Spec.Id = "some-other-stack-id"
			expectedStack.Spec.BuildImage.Image = "some-uploaded-build-image-2@some-digest"
			expectedStack.Spec.RunImage.Image = "some-uploaded-run-image-2@some-digest"

			expectedDefaultStack := defaultStack.DeepCopy()
			expectedDefaultStack.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultStack.Spec.Id = "some-other-stack-id"
			expectedDefaultStack.Spec.BuildImage.Image = "some-uploaded-build-image-2@some-digest"
			expectedDefaultStack.Spec.RunImage.Image = "some-uploaded-run-image-2@some-digest"

			expectedBuilder := builder.DeepCopy()
			expectedBuilder.Annotations[importTimestampKey] = newTimestamp
			expectedBuilder.Spec.Order = []v1alpha1.OrderEntry{
				{
					Group: []v1alpha1.BuildpackRef{
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id: "buildpack-2",
							},
						},
					},
				},
			}

			expectedDefaultBuilder := defaultBuilder.DeepCopy()
			expectedDefaultBuilder.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultBuilder.Spec.Order = []v1alpha1.OrderEntry{
				{
					Group: []v1alpha1.BuildpackRef{
						{
							BuildpackInfo: v1alpha1.BuildpackInfo{
								Id: "buildpack-2",
							},
						},
					},
				},
			}

			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-2"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-2"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			it("creates stores, stacks, and cbs defined in the dependency descriptor and updates the timestamp", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/updated-deps.yaml",
					},
					ExpectedOutput: `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'...
	Added Buildpackage
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
Importing ClusterBuilder 'default'...
Imported resources
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedStore,
						},
						{
							Object: expectedStack,
						},
						{
							Object: expectedDefaultStack,
						},
						{
							Object: expectedBuilder,
						},
						{
							Object: expectedDefaultBuilder,
						},
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	it("errors when the apiVersion is unexpected", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{},
			Args: []string{
				"-f", "./testdata/invalid-deps.yaml",
			},
			ExpectedOutput: "Error: did not find expected apiVersion, must be one of: [kp.kpack.io/v1alpha1 kp.kpack.io/v1alpha2]\n",
			ExpectErr:      true,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		const expectedOutput = `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'...
Importing ClusterStack 'some-stack'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'...
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'...
Importing ClusterBuilder 'default'...
`

		builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
		defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-store","creationTimestamp":null},"spec":{"sources":[{"image":"new-registry.io/new-project/store-image@sha256:123abc"}]},"status":{}}'
  creationTimestamp: null
  name: some-store
spec:
  sources:
  - image: new-registry.io/new-project/store-image@sha256:123abc
status: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-uploaded-build-image@some-digest
  id: some-stack-id
  runImage:
    image: some-uploaded-run-image@some-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: default
spec:
  buildImage:
    image: some-uploaded-build-image@some-digest
  id: some-stack-id
  runImage:
    image: some-uploaded-run-image@some-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: some-cb
spec:
  order:
  - group:
    - id: buildpack-1
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: new-registry.io/new-project/some-cb
status:
  stack: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: default
spec:
  order:
  - group:
    - id: buildpack-1
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: new-registry.io/new-project/default
status:
  stack: {}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
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
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-store",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterStore\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"some-store\",\"creationTimestamp\":null},\"spec\":{\"sources\":[{\"image\":\"new-registry.io/new-project/store-image@sha256:123abc\"}]},\"status\":{}}"
        }
    },
    "spec": {
        "sources": [
            {
                "image": "new-registry.io/new-project/store-image@sha256:123abc"
            }
        ]
    },
    "status": {}
}
{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-stack",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z"
        }
    },
    "spec": {
        "id": "some-stack-id",
        "buildImage": {
            "image": "some-uploaded-build-image@some-digest"
        },
        "runImage": {
            "image": "some-uploaded-run-image@some-digest"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "default",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z"
        }
    },
    "spec": {
        "id": "some-stack-id",
        "buildImage": {
            "image": "some-uploaded-build-image@some-digest"
        },
        "runImage": {
            "image": "some-uploaded-run-image@some-digest"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
{
    "kind": "ClusterBuilder",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-cb",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"some-cb\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"new-registry.io/new-project/some-cb\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"buildpack-1\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "new-registry.io/new-project/some-cb",
        "stack": {
            "kind": "ClusterStack",
            "name": "some-stack"
        },
        "store": {
            "kind": "ClusterStore",
            "name": "some-store"
        },
        "order": [
            {
                "group": [
                    {
                        "id": "buildpack-1"
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
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "default",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/import-timestamp": "2006-01-02T15:04:05Z",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"new-registry.io/new-project/default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"some-stack\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"some-store\"},\"order\":[{\"group\":[{\"id\":\"buildpack-1\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"
        }
    },
    "spec": {
        "tag": "new-registry.io/new-project/default",
        "stack": {
            "kind": "ClusterStack",
            "name": "some-stack"
        },
        "store": {
            "kind": "ClusterStore",
            "name": "some-store"
        },
        "order": [
            {
                "group": [
                    {
                        "id": "buildpack-1"
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
				K8sObjects: []runtime.Object{
					config,
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
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
		defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		it("does not create any resources and prints result with dry run indicated", func() {
			const expectedOutput = `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'... (dry run)
Importing ClusterStack 'some-stack'... (dry run)
Importing ClusterStack 'default'... (dry run)
Importing ClusterBuilder 'some-cb'... (dry run)
Importing ClusterBuilder 'default'... (dry run)
Imported resources (dry run)
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
					"--dry-run",
				},
				ExpectedOutput: expectedOutput,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-store","creationTimestamp":null},"spec":{"sources":[{"image":"new-registry.io/new-project/store-image@sha256:123abc"}]},"status":{}}'
  creationTimestamp: null
  name: some-store
spec:
  sources:
  - image: new-registry.io/new-project/store-image@sha256:123abc
status: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-uploaded-build-image@some-digest
  id: some-stack-id
  runImage:
    image: some-uploaded-run-image@some-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
  creationTimestamp: null
  name: default
spec:
  buildImage:
    image: some-uploaded-build-image@some-digest
  id: some-stack-id
  runImage:
    image: some-uploaded-run-image@some-digest
status:
  buildImage: {}
  runImage: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-cb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-cb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: some-cb
spec:
  order:
  - group:
    - id: buildpack-1
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: new-registry.io/new-project/some-cb
status:
  stack: {}
---
apiVersion: kpack.io/v1alpha1
kind: ClusterBuilder
metadata:
  annotations:
    kpack.io/import-timestamp: "2006-01-02T15:04:05Z"
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}'
  creationTimestamp: null
  name: default
spec:
  order:
  - group:
    - id: buildpack-1
  serviceAccountRef:
    name: some-serviceaccount
    namespace: kpack
  stack:
    kind: ClusterStack
    name: some-stack
  store:
    kind: ClusterStore
    name: some-store
  tag: new-registry.io/new-project/default
status:
  stack: {}
`

			const expectedOutput = `Changes

ClusterStores

some-diff

ClusterStacks

some-diff

some-diff

ClusterBuilders

some-diff

some-diff


Importing ClusterStore 'some-store'... (dry run)
Importing ClusterStack 'some-stack'... (dry run)
Uploading to 'new-registry.io/new-project'...
Importing ClusterStack 'default'... (dry run)
Uploading to 'new-registry.io/new-project'...
Importing ClusterBuilder 'some-cb'... (dry run)
Importing ClusterBuilder 'default'... (dry run)
`

			it("does not create a Builder and prints the resource output", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
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
}

type FakeTimestampProvider struct {
	timestamp string
}

func (f FakeTimestampProvider) GetTimestamp() string {
	return f.timestamp
}
