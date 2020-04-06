package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/image"
)

func TestImageLister(t *testing.T) {
	spec.Run(t, "TestImageLister", testImageLister)
}

func testImageLister(t *testing.T, when spec.G, it spec.S) {
	var (
		lister = &image.Lister{
			DefaultNamespace: defaultNamespace,
			KpackClient:      fake.NewSimpleClientset(),
		}
	)

	when("there are images in the namespace", func() {
		lister.KpackClient = fake.NewSimpleClientset(&v1alpha1.ImageList{
			Items: []v1alpha1.Image{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image",
						Namespace: "other-namespace",
					},
				},
			},
		})

		it("returns the image list", func() {
			imageList, err := lister.List("other-namespace")
			require.NoError(t, err)
			require.Len(t, imageList.Items, 1)
			require.Equal(t, "test-image", imageList.Items[0].Name)
			require.Equal(t, "other-namespace", imageList.Items[0].Namespace)
		})
	})

	when("there are no images in the namespace", func() {
		it("returns an empty image list", func() {
			imageList, err := lister.List("empty-namespace")
			require.NoError(t, err)
			require.Len(t, imageList.Items, 0)
		})
	})

	when("the provided namespace is empty", func() {
		lister.KpackClient = fake.NewSimpleClientset(&v1alpha1.ImageList{
			Items: []v1alpha1.Image{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image",
						Namespace: defaultNamespace,
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-3",
						Namespace: "bar",
					},
				},
			},
		})

		it("uses the default namespace", func() {
			imageList, err := lister.List("")
			require.NoError(t, err)
			require.Len(t, imageList.Items, 1)
			require.Equal(t, "test-image", imageList.Items[0].Name)
			require.Equal(t, defaultNamespace, imageList.Items[0].Namespace)
		})
	})
}
