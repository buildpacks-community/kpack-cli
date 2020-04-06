package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

const defaultNamespace = "test-namespace"

func TestImageApplier(t *testing.T) {
	spec.Run(t, "TestImageApplier", testImageApplier)
}

func testImageApplier(t *testing.T, when spec.G, it spec.S) {
	var (
		imageConfig = &v1alpha1.Image{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-image",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha1.ImageSpec{
				Tag: "test-registry.io/test-tag",
				Source: v1alpha1.SourceConfig{
					Git: &v1alpha1.Git{
						URL:      "test-git-url",
						Revision: "test-git-revision",
					},
				},
			},
		}
	)

	when("the image does not exist", func() {
		it("the applier creates the image", func() {
			ApplyTest{
				ImageConfig: imageConfig,
				ExpectCreates: []runtime.Object{
					imageConfig,
				},
			}.test(t)
		})
	})

	when("the image exists", func() {
		it("the applier updates the image", func() {
			updatedImageConfig := imageConfig.DeepCopy()
			updatedImageConfig.Spec.Source.Git.Revision = "new-git-revision"

			ApplyTest{
				Objects: []runtime.Object{
					imageConfig,
				},
				ImageConfig: updatedImageConfig,
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: updatedImageConfig,
					},
				},
			}.test(t)
		})
	})

	when("the namespace is not specified in the image config", func() {
		it("uses the default namespace", func() {
			configWithoutNS := imageConfig.DeepCopy()
			configWithoutNS.Namespace = ""

			ApplyTest{
				ImageConfig: configWithoutNS,
				ExpectCreates: []runtime.Object{
					imageConfig,
				},
			}.test(t)
		})
	})
}

type ApplyTest struct {
	Objects       []runtime.Object
	ImageConfig   *v1alpha1.Image
	ExpectUpdates []clientgotesting.UpdateActionImpl
	ExpectCreates []runtime.Object
}

func (a ApplyTest) test(t *testing.T) {
	t.Helper()
	client := fake.NewSimpleClientset(a.Objects...)

	applier := &image.Applier{
		DefaultNamespace: defaultNamespace,
		KpackClient:      client,
	}
	err := applier.Apply(a.ImageConfig)
	require.NoError(t, err)

	testhelpers.TestUpdatesAndCreates(t, client, a.ExpectUpdates, a.ExpectCreates)
}
