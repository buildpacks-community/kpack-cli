package kpackcompat

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1alpha1Client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// images implements ImageInterface
type images struct {
	client v1alpha1Client.KpackV1alpha1Interface
	ns     string
}

// newImages returns a Images
func newImages(c *kpackV1alpha1CompatClient, namespace string) *images {
	return &images{
		client: c.v1alpha1KpackClient,
		ns:     namespace,
	}
}

func (i *images) Create(ctx context.Context, image *v1alpha2.Image, opts v1.CreateOptions) (*v1alpha2.Image, error) {
	convertedImage, err := convertToV1Image(ctx, image)
	if err != nil {
		return nil, err
	}

	createdV1Image, err := i.client.Images(i.ns).Create(ctx, convertedImage, opts)
	if err != nil {
		return nil, err
	}

	createdV2Image, err := convertFromV1Image(ctx, createdV1Image)
	if err != nil {
		return nil, err
	}

	return createdV2Image, nil
}

func (i *images) Update(ctx context.Context, image *v1alpha2.Image, opts v1.UpdateOptions) (*v1alpha2.Image, error) {
	convertedImage, err := convertToV1Image(ctx, image)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := i.client.Images(i.ns).Update(ctx, convertedImage, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Image, err := convertFromV1Image(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2Image, nil
}

func (i *images) UpdateStatus(ctx context.Context, image *v1alpha2.Image, opts v1.UpdateOptions) (*v1alpha2.Image, error) {
	convertedImage, err := convertToV1Image(ctx, image)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := i.client.Images(i.ns).UpdateStatus(ctx, convertedImage, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Image, err := convertFromV1Image(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2Image, nil
}

func (i *images) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	err := i.client.Images(i.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (i *images) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	err := i.client.Images(i.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (i *images) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.Image, error) {
	v1Image, err := i.client.Images(i.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedImage, err := convertFromV1Image(ctx, v1Image)
	if err != nil {
		return nil, err
	}
	return convertedImage, nil
}

func (i *images) List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.ImageList, error) {
	compatList, err := i.client.Images(i.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.ImageList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.Image{},
	}

	for _, compatObj := range compatList.Items {
		convertedImage, err := convertFromV1Image(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedImage)
	}

	return list, nil
}

func (i *images) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	watchable, err := i.client.Images(i.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (i *images) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Image, err error) {
	v1Result, err := i.client.Images(i.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := convertFromV1Image(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func convertFromV1Image(ctx context.Context, v1Image *v1alpha1.Image) (*v1alpha2.Image, error) {
	resultImage := v1alpha2.Image{}
	err := resultImage.ConvertFrom(ctx, v1Image)
	if err != nil {
		return nil, err
	}
	return &resultImage, nil
}

func convertToV1Image(ctx context.Context, v2Image *v1alpha2.Image) (*v1alpha1.Image, error) {
	resultImage := &v1alpha1.Image{}
	err := v2Image.ConvertTo(ctx, resultImage)
	if err != nil {
		return nil, err
	}
	return resultImage, nil
}
