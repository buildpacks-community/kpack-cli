package kpack

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	buildV1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1alpha1Client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	v1alpha2Client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

type ImagesGetter interface {
	Images(namespace string) v1alpha2Client.ImageInterface
}

// images implements ImageInterface
type images struct {
	client rest.Interface
	ns     string
}

// newImages returns a Images
func newImages(c *KpackV1alpha2Client, namespace string) *images {
	return &images{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *images) Create(ctx context.Context, image *buildV1alpha2.Image, opts v1.CreateOptions) (*buildV1alpha2.Image, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedImage, err := c.convertToV1Image(ctx, image)
	if err != nil {
		return nil, err
	}

	createdV1Image, err := v1Client.Images(c.ns).Create(ctx, convertedImage, opts)
	if err != nil {
		return nil, err
	}

	createdV2Image, err := c.convertFromV1Image(ctx, createdV1Image)
	if err != nil {
		return nil, err
	}

	return createdV2Image, nil
}

func (c *images) Update(ctx context.Context, image *buildV1alpha2.Image, opts v1.UpdateOptions) (*buildV1alpha2.Image, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedImage, err := c.convertToV1Image(ctx, image)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := v1Client.Images(c.ns).Update(ctx, convertedImage, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Image, err := c.convertFromV1Image(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2Image, nil
}

func (c *images) UpdateStatus(ctx context.Context, image *buildV1alpha2.Image, opts v1.UpdateOptions) (*buildV1alpha2.Image, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedImage, err := c.convertToV1Image(ctx, image)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := v1Client.Images(c.ns).UpdateStatus(ctx, convertedImage, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Image, err := c.convertFromV1Image(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2Image, nil
}

func (c *images) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.Images(c.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (c *images) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.Images(c.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (c *images) Get(ctx context.Context, name string, opts v1.GetOptions) (*buildV1alpha2.Image, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Image, err := v1Client.Images(c.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedImage, err := c.convertFromV1Image(ctx, v1Image)
	if err != nil {
		return nil, err
	}
	return convertedImage, nil
}

func (c *images) List(ctx context.Context, opts v1.ListOptions) (*buildV1alpha2.ImageList, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}
	compatList, err := v1Client.Images(c.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.ImageList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.Image{},
	}

	for _, compatObj := range compatList.Items {
		convertedImage, err := c.convertFromV1Image(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedImage)
	}

	return list, nil
}

func (c *images) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	watchable, err := v1Client.Images(c.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (c *images) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *buildV1alpha2.Image, err error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Result, err := v1Client.Images(c.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := c.convertFromV1Image(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func (c *images) convertFromV1Image(ctx context.Context, v1Image *v1alpha1.Image) (result *buildV1alpha2.Image, err error) {
	resultImage := buildV1alpha2.Image{}
	err = resultImage.ConvertFrom(ctx, v1Image)
	if err != nil {
		return nil, err
	}
	return &resultImage, nil
}

func (c *images) convertToV1Image(ctx context.Context, v2Image *buildV1alpha2.Image) (result *v1alpha1.Image, err error) {
	resultImage := v1alpha1.Image{}
	err = resultImage.ConvertTo(ctx, v2Image)
	if err != nil {
		return nil, err
	}
	return &resultImage, nil
}
