package _import_test

import (
	"fmt"
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
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	storefakes "github.com/pivotal/build-service-cli/pkg/clusterstore/fakes"
	importcmds "github.com/pivotal/build-service-cli/pkg/commands/import"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImportCommand(t *testing.T) {
	spec.Run(t, "TestImportCommand", testImportCommand)
}

func testImportCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		importTimestampKey = "kpack.io/import-timestamp"
	)
	fakeBuildpackageUploader := storefakes.FakeBuildpackageUploader{
		"some-registry.io/some-project/store-image":   "new-registry.io/new-project/store-image@sha256:123abc",
		"some-registry.io/some-project/store-image-2": "new-registry.io/new-project/store-image-2@sha256:456def",
	}

	storeFactory := &clusterstore.Factory{
		Uploader: fakeBuildpackageUploader,
	}

	buildImage, buildImageId, runImage, runImageId := makeStackImages(t, "some-stack-id")
	buildImage2, buildImage2Id, runImage2, runImage2Id := makeStackImages(t, "some-other-stack-id")

	fetcher := &fakes.Fetcher{}
	fetcher.AddImage("some-registry.io/some-project/build-image", buildImage)
	fetcher.AddImage("some-registry.io/some-project/run-image", runImage)
	fetcher.AddImage("some-registry.io/some-project/build-image-2", buildImage2)
	fetcher.AddImage("some-registry.io/some-project/run-image-2", runImage2)

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
				Image: "new-registry.io/new-project/build@" + buildImageId,
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: "new-registry.io/new-project/run@" + runImageId,
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
			Name: "some-ccb",
			Annotations: map[string]string{
				importTimestampKey: timestampProvider.timestamp,
			},
		},
		Spec: v1alpha1.ClusterBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: "new-registry.io/new-project/some-ccb",
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

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return importcmds.NewImportCommand(clientSetProvider, timestampProvider, storeFactory, stackFactory)
	}

	when("there are no stores, stacks, or ccbs", func() {
		it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"ClusterBuilder","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"-f", "./testdata/deps.yaml",
				},
				ExpectedOutput: "Importing Cluster Store 'some-store'...\nUploading to 'new-registry.io/new-project'...\nImporting Cluster Stack 'some-stack'...\nImporting Cluster Stack 'default'...\nImporting Cluster Builder 'some-ccb'...\nImporting Cluster Builder 'default'...\n",
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

	when("there are existing stores, stacks, or ccbs", func() {
		when("the dependency descriptor and the store have the exact same objects", func() {
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

			it("updates the import timestamp", func() {
				stack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImageId)
				stack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImageId)

				defaultStack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImageId)
				defaultStack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImageId)

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
					ExpectedOutput: "Importing Cluster Store 'some-store'...\nUploading to 'new-registry.io/new-project'...\nBuildpackage 'new-registry.io/new-project/store-image@sha256:123abc' already exists in the store\nImporting Cluster Stack 'some-stack'...\nImporting Cluster Stack 'default'...\nImporting Cluster Builder 'some-ccb'...\nImporting Cluster Builder 'default'...\n",
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

			expectedStore := store.DeepCopy()
			expectedStore.Annotations[importTimestampKey] = newTimestamp
			expectedStore.Spec.Sources = append(expectedStore.Spec.Sources, v1alpha1.StoreImage{
				Image: "new-registry.io/new-project/store-image-2@sha256:456def",
			})

			expectedStack := stack.DeepCopy()
			expectedStack.Annotations[importTimestampKey] = newTimestamp
			expectedStack.Spec.Id = "some-other-stack-id"
			expectedStack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImage2Id)
			expectedStack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImage2Id)

			expectedDefaultStack := defaultStack.DeepCopy()
			expectedDefaultStack.Annotations[importTimestampKey] = newTimestamp
			expectedDefaultStack.Spec.Id = "some-other-stack-id"
			expectedDefaultStack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImage2Id)
			expectedDefaultStack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImage2Id)

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

			it("creates stores, stacks, and ccbs defined in the dependency descriptor and updates the timestamp", func() {
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
					ExpectedOutput: "Importing Cluster Store 'some-store'...\nUploading to 'new-registry.io/new-project'...\nAdded Buildpackage 'new-registry.io/new-project/store-image-2@sha256:456def'\nImporting Cluster Stack 'some-stack'...\nImporting Cluster Stack 'default'...\nImporting Cluster Builder 'some-ccb'...\nImporting Cluster Builder 'default'...\n",
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

type FakeTimestampProvider struct {
	timestamp string
}

func (f FakeTimestampProvider) GetTimestamp() string {
	return f.timestamp
}
