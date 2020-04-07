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

func TestImageDeleter(t *testing.T) {
	spec.Run(t, "TestImageDeleter", testImageDeleter)
}

func testImageDeleter(t *testing.T, when spec.G, it spec.S) {
	when("the image exists", func() {
		it("the deleter deletes the image", func() {
			DeleterTest{
				Namespace: "test-namespace",
				Name:      "test-image",
				Objects: []runtime.Object{
					&v1alpha1.Image{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-image",
							Namespace: "test-namespace",
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
					},
				},
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						ActionImpl: clientgotesting.ActionImpl{
							Namespace: "test-namespace",
						},
						Name: "test-image",
					},
				},
			}.test(t)
		})
	})

	when("the image does not exist", func() {
		it("the deleter returns an error", func() {
			DeleterTest{
				Namespace:   "test-namespace",
				Name:        "test-image",
				ExpectError: true,
			}.test(t)
		})
	})
}

type DeleterTest struct {
	Namespace     string
	Name          string
	Objects       []runtime.Object
	ExpectDeletes []clientgotesting.DeleteActionImpl
	ExpectError   bool
}

func (d DeleterTest) test(t *testing.T) {
	t.Helper()
	client := fake.NewSimpleClientset(d.Objects...)

	deleter := &image.Deleter{
		KpackClient: client,
	}
	err := deleter.Delete(d.Namespace, d.Name)
	if d.ExpectError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)

	testhelpers.TestDeletes(t, client, d.ExpectDeletes)
}
