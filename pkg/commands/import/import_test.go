package _import_test

import (
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
		"some-registry.io/some-project/store-image": "new-registry.io/new-project/store-image",
	}

	storeFactory := &storepkg.Factory{
		Uploader: fakeBuildpackageUploader,
	}

	buildImage, buildImageId, runImage, runImageId := makeStackImages(t, "some-stack-id")

	fetcher := &fakes.Fetcher{}
	fetcher.AddImage("some-registry.io/some-project/build-image", buildImage)
	fetcher.AddImage("some-registry.io/some-project/run-image", runImage)

	relocator := &fakes.Relocator{}

	stackFactory := &stackpkg.Factory{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

	expectedStore := &expv1alpha1.Store{
		TypeMeta: metav1.TypeMeta{
			Kind:       expv1alpha1.StoreKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-store",
			Annotations: map[string]string{
				"buildservice.pivotal.io/defaultRepository":        "new-registry.io/new-project",
				"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Store","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-store","creationTimestamp":null,"annotations":{"buildservice.pivotal.io/defaultRepository":"new-registry.io/new-project"}},"spec":{"sources":[{"image":"new-registry.io/new-project/store-image"}]},"status":{}}`,
			},
		},
		Spec: expv1alpha1.StoreSpec{
			Sources: []expv1alpha1.StoreImage{
				{Image: "new-registry.io/new-project/store-image"},
			},
		},
	}

	expectedStack := &expv1alpha1.Stack{
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

	defaultStack := expectedStack.DeepCopy()
	defaultStack.Name = "default"

	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ccb-secret-",
			Namespace:    "kpack",
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(`{"auths":{"new-registry.io":{"username":"some-user","password":"some-password"}}}`),
		},
		Type: "kubernetes.io/dockerconfigjson",
	}

	expectedServiceAccount := &corev1.ServiceAccount{
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

	expectedBuilder := &expv1alpha1.CustomClusterBuilder{
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
									Id: "some-registry.io/some-project/buildpackage",
								},
							},
						},
					},
				},
			},
		},
	}

	defaultBuilder := expectedBuilder.DeepCopy()
	defaultBuilder.Name = "default"
	defaultBuilder.Spec.Tag = "new-registry.io/new-project/default"

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		k8sClientSet.PrependReactor("create", "secrets", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			secret := expectedSecret.DeepCopy()
			secret.Name = secret.GenerateName + "test"
			return true, secret, nil
		})

		k8sClientSet.PrependReactor("create", "serviceaccounts", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			sa := expectedServiceAccount.DeepCopy()
			sa.Name = sa.GenerateName + "test"
			return true, sa, nil
		})

		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return importcmds.NewImportCommand(clientSetProvider, storeFactory, stackFactory)
	}

	when("there are no stores, stacks, or ccbs", func() {
		when("a username and password are provided", func() {
			expectedBuilder.Spec.ServiceAccountRef = corev1.ObjectReference{
				Namespace: "kpack",
				Name:      "ccb-serviceaccount-test",
			}
			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{"namespace":"kpack","name":"ccb-serviceaccount-test"}},"status":{"stack":{}}}`

			defaultBuilder.Spec.ServiceAccountRef = corev1.ObjectReference{
				Namespace: "kpack",
				Name:      "ccb-serviceaccount-test",
			}
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null,"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"CustomClusterBuilder\",\"apiVersion\":\"experimental.kpack.pivotal.io/v1alpha1\",\"metadata\":{\"name\":\"some-ccb\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"new-registry.io/new-project/some-ccb\",\"stack\":\"some-stack\",\"store\":\"some-store\",\"order\":[{\"group\":[{\"id\":\"some-registry.io/some-project/buildpackage\"}]}],\"serviceAccountRef\":{\"namespace\":\"kpack\",\"name\":\"ccb-serviceaccount-test\"}},\"status\":{\"stack\":{}}}"}},"spec":{"tag":"new-registry.io/new-project/default","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{"namespace":"kpack","name":"ccb-serviceaccount-test"}},"status":{"stack":{}}}`

			it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
				testhelpers.CommandTest{
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"-r", "new-registry.io/new-project",
						"-u", "some-user",
						"-p", "some-password",
					},
					ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\n",
					ExpectCreates: []runtime.Object{
						expectedSecret,
						expectedServiceAccount,
						expectedStore,
						expectedStack,
						defaultStack,
						expectedBuilder,
						defaultBuilder,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("a username and password are not provided", func() {
			expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"some-ccb","creationTimestamp":null},"spec":{"tag":"new-registry.io/new-project/some-ccb","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`
			defaultBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"default","creationTimestamp":null,"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"CustomClusterBuilder\",\"apiVersion\":\"experimental.kpack.pivotal.io/v1alpha1\",\"metadata\":{\"name\":\"some-ccb\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"new-registry.io/new-project/some-ccb\",\"stack\":\"some-stack\",\"store\":\"some-store\",\"order\":[{\"group\":[{\"id\":\"some-registry.io/some-project/buildpackage\"}]}],\"serviceAccountRef\":{}},\"status\":{\"stack\":{}}}"}},"spec":{"tag":"new-registry.io/new-project/default","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"some-registry.io/some-project/buildpackage"}]}],"serviceAccountRef":{}},"status":{"stack":{}}}`

			it("creates stores, stacks, and ccbs defined in the dependency descriptor", func() {
				testhelpers.CommandTest{
					Args: []string{
						"-f", "./testdata/deps.yaml",
						"-r", "new-registry.io/new-project",
					},
					ExpectedOutput: "Uploading to 'new-registry.io/new-project'...\n",
					ExpectCreates: []runtime.Object{
						expectedStore,
						expectedStack,
						defaultStack,
						expectedBuilder,
						defaultBuilder,
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
