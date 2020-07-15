package _import_test

import (
	"fmt"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	importcmds "github.com/pivotal/build-service-cli/pkg/commands/import"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	stackpkg "github.com/pivotal/build-service-cli/pkg/stack"
	storepkg "github.com/pivotal/build-service-cli/pkg/store"
	storefakes "github.com/pivotal/build-service-cli/pkg/store/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImportCommand(t *testing.T) {
	spec.Run(t, "TestImportCommand", testImportCommand)
}

func testImportCommand(t *testing.T, when spec.G, it spec.S) {
	fakeBuildpackageUploader := storefakes.FakeBuildpackageUploader{
		"some-registry.io/some-project/store-image":   "new-registry.io/new-project/store-image@sha256:123abc",
		"some-registry.io/some-project/store-image-2": "new-registry.io/new-project/store-image-2@sha256:456def",
	}

	storeFactory := &storepkg.Factory{
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

	stackFactory := &stackpkg.Factory{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

	store := &expv1alpha1.Store{
		TypeMeta: metav1.TypeMeta{
			Kind:       expv1alpha1.StoreKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-store",
			Annotations: map[string]string{
				"buildservice.pivotal.io/defaultRepository":        "new-registry.io/new-project",
				"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Store","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-store","creationTimestamp":null,"annotations":{"buildservice.pivotal.io/defaultRepository":"new-registry.io/new-project"}},"spec":{"sources":[{"image":"new-registry.io/new-project/store-image@sha256:123abc"}]},"status":{}}`,
			},
		},
		Spec: expv1alpha1.StoreSpec{
			Sources: []expv1alpha1.StoreImage{
				{Image: "new-registry.io/new-project/store-image@sha256:123abc"},
			},
		},
	}

	stack := &expv1alpha1.Stack{
		TypeMeta: metav1.TypeMeta{
			Kind:       expv1alpha1.StackKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-stack",
			Annotations: map[string]string{
				stackpkg.DefaultRepositoryAnnotation: "new-registry.io/new-project",
			},
		},
		Spec: expv1alpha1.StackSpec{
			Id: "some-stack-id",
			BuildImage: expv1alpha1.StackSpecImage{
				Image: "new-registry.io/new-project/build@" + buildImageId,
			},
			RunImage: expv1alpha1.StackSpecImage{
				Image: "new-registry.io/new-project/run@" + runImageId,
			},
		},
	}

	defaultStack := stack.DeepCopy()
	defaultStack.Name = "default"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ccb-secret-",
			Namespace:    "kpack",
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(`{"auths":{"new-registry.io":{"username":"some-user","password":"some-password"}}}`),
		},
		Type: "kubernetes.io/dockerconfigjson",
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ccb-serviceaccount-",
			Namespace:    "kpack",
		},
		Secrets: []corev1.ObjectReference{
			{
				Namespace: "kpack",
				Name:      "ccb-secret-test",
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "ccb-secret-test",
			},
		},
	}

	builder := &expv1alpha1.CustomClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       expv1alpha1.CustomClusterBuilderKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "some-ccb",
			Annotations: map[string]string{},
		},
		Spec: expv1alpha1.CustomClusterBuilderSpec{
			CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
				Tag:   "new-registry.io/new-project/some-ccb",
				Stack: "some-stack",
				Store: "some-store",
				Order: []expv1alpha1.OrderEntry{
					{
						Group: []expv1alpha1.BuildpackRef{
							{
								BuildpackInfo: expv1alpha1.BuildpackInfo{
									Id: "buildpack-1",
								},
							},
						},
					},
				},
			},
		},
	}

	defaultBuilder := builder.DeepCopy()
	defaultBuilder.Name = "default"
	defaultBuilder.Spec.Tag = "new-registry.io/new-project/default"

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		k8sClientSet.PrependReactor("create", "secrets", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			secret := secret.DeepCopy()
			secret.Name = secret.GenerateName + "test"
			return true, secret, nil
		})

		k8sClientSet.PrependReactor("create", "serviceaccounts", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			sa := serviceAccount.DeepCopy()
			sa.Name = sa.GenerateName + "test"
			return true, sa, nil
		})

		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return importcmds.NewImportCommand(clientSetProvider, storeFactory, stackFactory)
	}

	when("there are no stores, stacks, or ccbs", func() {
		when("a username and password are provided", func() {
			it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
				builder.Spec.ServiceAccountRef = corev1.ObjectReference{
					Namespace: "kpack",
					Name:      "ccb-serviceaccount-test",
				}
				builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"ccb-serviceaccount-test"}},"status":{"stack":{}}}`

				defaultBuilder.Spec.ServiceAccountRef = corev1.ObjectReference{
					Namespace: "kpack",
					Name:      "ccb-serviceaccount-test",
				}
				defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{"namespace":"kpack","name":"ccb-serviceaccount-test"}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"-r", "new-registry.io/new-project",
						"-u", "some-user",
						"-p", "some-password",
					},
					ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\n",
					ExpectCreates: []runtime.Object{
						secret,
						serviceAccount,
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("a username and password are not provided", func() {
			it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
				builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`
				defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/default","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"buildpack-1"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`

				testhelpers.CommandTest{
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"-r", "new-registry.io/new-project",
					},
					ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\n",
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
	})

	when("there are existing stores, stacks, or ccbs", func() {
		when("the dependency descriptor and the store have the exact same objects", func() {
			it("does not change any of the existing resources", func() {
				stack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImageId)
				stack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImageId)

				defaultStack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImageId)
				defaultStack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImageId)

				testhelpers.CommandTest{
					KpackObjects: []runtime.Object{
						store,
						stack,
						defaultStack,
						builder,
						defaultBuilder,
					},
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"-r", "new-registry.io/new-project",
					},
					ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\nBuildpackage 'new-registry.io/new-project/store-image@sha256:123abc' already exists in the store\n",
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("the dependency descriptor has different resources", func() {
			builder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null,"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"CustomClusterBuilder\",\"apiVersion\":\"experimental.kpack.pivotal.io/v1alpha1\",\"metadata\":{\"name\":\"some-ccb\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"new-registry.io/new-project/some-ccb\",\"stack\":\"some-stack\",\"store\":\"some-store\",\"order\":[{\"group\":[{\"id\":\"some-registry.io/some-project/buildpackage\"}]}],\"serviceAccountRef\":{}},\"status\":{\"stack\":{}}}"}},"spec":{"tag":"new-registry.io/new-project/default","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`

			expectedStore := store.DeepCopy()
			expectedStore.Spec.Sources = append(expectedStore.Spec.Sources, expv1alpha1.StoreImage{
				Image: "new-registry.io/new-project/store-image-2@sha256:456def",
			})

			expectedStack := stack.DeepCopy()
			expectedStack.Spec.Id = "some-other-stack-id"
			expectedStack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImage2Id)
			expectedStack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImage2Id)

			expectedDefaultStack := defaultStack.DeepCopy()
			expectedDefaultStack.Spec.Id = "some-other-stack-id"
			expectedDefaultStack.Spec.BuildImage.Image = fmt.Sprintf("new-registry.io/new-project/build@%s", buildImage2Id)
			expectedDefaultStack.Spec.RunImage.Image = fmt.Sprintf("new-registry.io/new-project/run@%s", runImage2Id)

			expectedBuilder := builder.DeepCopy()
			expectedBuilder.Spec.Order = []expv1alpha1.OrderEntry{
				{
					Group: []expv1alpha1.BuildpackRef{
						{
							BuildpackInfo: expv1alpha1.BuildpackInfo{
								Id: "buildpack-2",
							},
						},
					},
				},
			}

			expectedDefaultBuilder := defaultBuilder.DeepCopy()
			expectedDefaultBuilder.Spec.Order = []expv1alpha1.OrderEntry{
				{
					Group: []expv1alpha1.BuildpackRef{
						{
							BuildpackInfo: expv1alpha1.BuildpackInfo{
								Id: "buildpack-2",
							},
						},
					},
				},
			}

			when("a username and password are not provided", func() {
				it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
					testhelpers.CommandTest{
						KpackObjects: []runtime.Object{
							store,
							stack,
							defaultStack,
							builder,
							defaultBuilder,
						},
						Args: []string{
							"-f", "./testdata/updated-deps.yaml",
							"-r", "new-registry.io/new-project",
						},
						ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\nAdded Buildpackage 'new-registry.io/new-project/store-image-2@sha256:456def'\n",
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

			when("a username and password are provided", func() {
				expectedSecret := secret.DeepCopy()

				exepectedServiceAccount := serviceAccount.DeepCopy()

				expectedBuilder.Spec.ServiceAccountRef = corev1.ObjectReference{
					Namespace: "kpack",
					Name:      "ccb-serviceaccount-test",
				}
				expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`

				expectedDefaultBuilder.Spec.ServiceAccountRef = corev1.ObjectReference{
					Namespace: "kpack",
					Name:      "ccb-serviceaccount-test",
				}
				expectedDefaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null,"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"CustomClusterBuilder\",\"apiVersion\":\"experimental.kpack.pivotal.io/v1alpha1\",\"metadata\":{\"name\":\"some-ccb\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"new-registry.io/new-project/some-ccb\",\"stack\":\"some-stack\",\"store\":\"some-store\",\"order\":[{\"group\":[{\"id\":\"some-registry.io/some-project/buildpackage\"}]}],\"serviceAccountRef\":{}},\"status\":{\"stack\":{}}}"}},"spec":{"tag":"new-registry.io/new-project/default","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`

				it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							secret,
							serviceAccount,
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
							"-r", "new-registry.io/new-project",
							"-u", "some-user",
							"-p", "some-password",
						},
						ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\nAdded Buildpackage 'new-registry.io/new-project/store-image-2@sha256:456def'\n",
						ExpectCreates: []runtime.Object{
							expectedSecret,
							exepectedServiceAccount,
						},
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
	})
}

func makeStackImages(t *testing.T, stackId string) (v1.Image, string, v1.Image, string) {
	buildImage, err := random.Image(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	buildImage, err = imagehelpers.SetStringLabel(buildImage, stackpkg.IdLabel, stackId)
	if err != nil {
		t.Fatal(err)
	}

	runImage, err := random.Image(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	runImage, err = imagehelpers.SetStringLabel(runImage, stackpkg.IdLabel, stackId)
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
