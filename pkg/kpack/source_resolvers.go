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

// sourceResolvers implements SourceResolverInterface
type sourceResolvers struct {
	client rest.Interface
	ns     string
}

// newSourceResolvers returns a SourceResolvers
func newSourceResolvers(c *KpackV1alpha2Client, namespace string) *sourceResolvers {
	return &sourceResolvers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *sourceResolvers) Create(ctx context.Context, sourceResolver *buildV1alpha2.SourceResolver, opts v1.CreateOptions) (*buildV1alpha2.SourceResolver, error) {
	v2Client := v1alpha2Client.KpackV1alpha2Client{}
	if c.client.APIVersion().String() == kpackGroupVersionV1alpha2 {
		sourceResolver, err := v2Client.SourceResolvers(c.ns).Create(ctx, sourceResolver, opts)
		if err != nil {
			return nil, err
		}
		return sourceResolver, nil
	}

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedSourceResolver, err := c.convertToV1SourceResolver(ctx, sourceResolver)
	if err != nil {
		return nil, err
	}

	createdV1SourceResolver, err := v1Client.SourceResolvers(c.ns).Create(ctx, convertedSourceResolver, opts)
	if err != nil {
		return nil, err
	}

	createdV2SourceResolver, err := c.convertFromV1SourceResolver(ctx, createdV1SourceResolver)
	if err != nil {
		return nil, err
	}

	return createdV2SourceResolver, nil
}

func (c *sourceResolvers) Update(ctx context.Context, sourceResolver *buildV1alpha2.SourceResolver, opts v1.UpdateOptions) (*buildV1alpha2.SourceResolver, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedSourceResolver, err := c.convertToV1SourceResolver(ctx, sourceResolver)
	if err != nil {
		return nil, err
	}

	updatedV1sourceResolver, err := v1Client.SourceResolvers(c.ns).Update(ctx, convertedSourceResolver, opts)
	if err != nil {
		return nil, err
	}

	updatedV2SourceResolver, err := c.convertFromV1SourceResolver(ctx, updatedV1sourceResolver)
	if err != nil {
		return nil, err
	}

	return updatedV2SourceResolver, nil
}

func (c *sourceResolvers) UpdateStatus(ctx context.Context, sourceResolver *buildV1alpha2.SourceResolver, opts v1.UpdateOptions) (*buildV1alpha2.SourceResolver, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedSourceResolver, err := c.convertToV1SourceResolver(ctx, sourceResolver)
	if err != nil {
		return nil, err
	}

	updatedV1sourceResolver, err := v1Client.SourceResolvers(c.ns).UpdateStatus(ctx, convertedSourceResolver, opts)
	if err != nil {
		return nil, err
	}

	updatedV2SourceResolver, err := c.convertFromV1SourceResolver(ctx, updatedV1sourceResolver)
	if err != nil {
		return nil, err
	}

	return updatedV2SourceResolver, nil
}

func (c *sourceResolvers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.SourceResolvers(c.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (c *sourceResolvers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.SourceResolvers(c.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (c *sourceResolvers) Get(ctx context.Context, name string, opts v1.GetOptions) (*buildV1alpha2.SourceResolver, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1SourceResolver, err := v1Client.SourceResolvers(c.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedSourceResolver, err := c.convertFromV1SourceResolver(ctx, v1SourceResolver)
	if err != nil {
		return nil, err
	}
	return convertedSourceResolver, nil
}

func (c *sourceResolvers) List(ctx context.Context, opts v1.ListOptions) (*buildV1alpha2.SourceResolverList, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}
	compatList, err := v1Client.SourceResolvers(c.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.SourceResolverList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.SourceResolver{},
	}

	for _, compatObj := range compatList.Items {
		convertedSourceResolver, err := c.convertFromV1SourceResolver(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedSourceResolver)
	}

	return list, nil
}

func (c *sourceResolvers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	watchable, err := v1Client.SourceResolvers(c.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (c *sourceResolvers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *buildV1alpha2.SourceResolver, err error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Result, err := v1Client.SourceResolvers(c.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := c.convertFromV1SourceResolver(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func (c *sourceResolvers) convertFromV1SourceResolver(ctx context.Context, v1SourceResolver *v1alpha1.SourceResolver) (result *buildV1alpha2.SourceResolver, err error) {
	resultSourceResolver := buildV1alpha2.SourceResolver{}
	err = resultSourceResolver.ConvertFrom(ctx, v1SourceResolver)
	if err != nil {
		return nil, err
	}
	return &resultSourceResolver, nil
}

func (c *sourceResolvers) convertToV1SourceResolver(ctx context.Context, v2SourceResolver *buildV1alpha2.SourceResolver) (result *v1alpha1.SourceResolver, err error) {
	resultSourceResolver := v1alpha1.SourceResolver{}
	err = resultSourceResolver.ConvertTo(ctx, v2SourceResolver)
	if err != nil {
		return nil, err
	}
	return &resultSourceResolver, nil
}
