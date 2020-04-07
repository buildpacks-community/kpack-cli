package image_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
)

func TestImageListCommand(t *testing.T) {
	spec.Run(t, "TestImageListCommand", testImageListCommand)
}

func testImageListCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		out         = &bytes.Buffer{}
		imageLister = newFakeImageLister()
		listCmd     = &image.ListCommand{
			Out:              out,
			Lister:           imageLister,
			DefaultNamespace: "default-namespace",
		}
	)

	when("a namespace is provided", func() {
		when("the namespaces has images", func() {
			imageLister.imageLists["test-namespace"] = &v1alpha1.ImageList{
				Items: []v1alpha1.Image{
					{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-image-1",
							Namespace: "test-namespace",
						},
						Status: v1alpha1.ImageStatus{
							Status: corev1alpha1.Status{
								Conditions: []corev1alpha1.Condition{
									{
										Type:   corev1alpha1.ConditionReady,
										Status: corev1.ConditionFalse,
									},
								},
							},
							LatestImage: "test-registry.io/test-image-1@sha256:abcdef123",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-image-2",
							Namespace: "test-namespace",
						},
						Status: v1alpha1.ImageStatus{
							Status: corev1alpha1.Status{
								Conditions: []corev1alpha1.Condition{
									{
										Type:   corev1alpha1.ConditionReady,
										Status: corev1.ConditionUnknown,
									},
								},
							},
							LatestImage: "test-registry.io/test-image-2@sha256:abcdef123",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-image-3",
							Namespace: "test-namespace",
						},
						Spec: v1alpha1.ImageSpec{},
						Status: v1alpha1.ImageStatus{
							Status: corev1alpha1.Status{
								Conditions: []corev1alpha1.Condition{
									{
										Type:   corev1alpha1.ConditionReady,
										Status: corev1.ConditionTrue,
									},
								},
							},
							LatestImage: "test-registry.io/test-image-3@sha256:abcdef123",
						},
					},
				},
			}

			it("returns a table of image details", func() {
				err := listCmd.Execute("test-namespace")
				require.NoError(t, err)

				expected := `Name            Ready      Latest Image
----            -----      ------------
test-image-1    False      test-registry.io/test-image-1@sha256:abcdef123
test-image-2    Unknown    test-registry.io/test-image-2@sha256:abcdef123
test-image-3    True       test-registry.io/test-image-3@sha256:abcdef123
`
				require.Equal(t, expected, out.String())
			})
		})

		when("the namespace has no images", func() {
			imageLister.imageLists["other-namespace"] = &v1alpha1.ImageList{}

			it("returns a message that the namespace has no images", func() {
				err := listCmd.Execute("other-namespace")
				require.NoError(t, err)

				require.Equal(t, "no images found in other-namespace namespace\n", out.String())
			})
		})
	})

	when("a namespace is not provided and the default namespace is used", func() {
		when("the namespace has no images", func() {
			imageLister.imageLists["default-namespace"] = &v1alpha1.ImageList{}

			it("returns a message that the namespace has no images", func() {
				err := listCmd.Execute("")
				require.NoError(t, err)

				require.Equal(t, "no images found in default-namespace namespace\n", out.String())
			})
		})
	})

	when("the lister returns an error", func() {
		imageLister.err = errors.New("some lister error")

		it("returns the listers error", func() {
			err := listCmd.Execute("test-namespace")
			require.Error(t, err, "some lister error")
		})
	})
}

type fakeImageLister struct {
	imageLists map[string]*v1alpha1.ImageList
	err        error
}

func newFakeImageLister() *fakeImageLister {
	return &fakeImageLister{
		imageLists: map[string]*v1alpha1.ImageList{},
	}
}

func (f *fakeImageLister) List(namespace string) (*v1alpha1.ImageList, error) {
	if f.err != nil {
		return nil, f.err
	}

	if imageList, ok := f.imageLists[namespace]; ok {
		return imageList, nil
	}
	return &v1alpha1.ImageList{}, nil
}
