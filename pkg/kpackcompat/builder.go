package kpackcompat

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1alpha1client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// builders implements BuilderInterface
type builders struct {
	client v1alpha1client.KpackV1alpha1Interface
	ns     string
}

func newBuilders(c *KpackV1alpha1CompatClient, namespace string) *builders {
	return &builders{
		client: c.v1alpha1KpackClient,
		ns:     namespace,
	}
}

func (b *builders) Create(ctx context.Context, builder *v1alpha2.Builder, opts metav1.CreateOptions) (*v1alpha2.Builder, error) {
	convertedBuilder, err := b.convertToV1Builder(ctx, builder)
	if err != nil {
		return nil, err
	}

	createdV1Builder, err := b.client.Builders(b.ns).Create(ctx, convertedBuilder, opts)
	if err != nil {
		return nil, err
	}

	createdV2Builder, err := b.convertFromV1Builder(ctx, createdV1Builder)
	if err != nil {
		return nil, err
	}

	return createdV2Builder, nil
}

func (b *builders) Update(ctx context.Context, builder *v1alpha2.Builder, opts metav1.UpdateOptions) (*v1alpha2.Builder, error) {
	convertedBuilder, err := b.convertToV1Builder(ctx, builder)
	if err != nil {
		return nil, err
	}

	updatedV1Builder, err := b.client.Builders(b.ns).Update(ctx, convertedBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Builder, err := b.convertFromV1Builder(ctx, updatedV1Builder)
	if err != nil {
		return nil, err
	}

	return updatedV2Builder, nil
}

func (b *builders) UpdateStatus(ctx context.Context, builder *v1alpha2.Builder, opts metav1.UpdateOptions) (*v1alpha2.Builder, error) {
	convertedBuilder, err := b.convertToV1Builder(ctx, builder)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := b.client.Builders(b.ns).UpdateStatus(ctx, convertedBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Builder, err := b.convertFromV1Builder(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2Builder, nil
}

func (b *builders) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := b.client.Builders(b.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (b *builders) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {

	err := b.client.Builders(b.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b *builders) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.Builder, error) {
	v1Builder, err := b.client.Builders(b.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedBuilder, err := b.convertFromV1Builder(ctx, v1Builder)
	if err != nil {
		return nil, err
	}
	return convertedBuilder, nil
}

func (b *builders) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.BuilderList, error) {
	compatList, err := b.client.Builders(b.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.BuilderList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.Builder{},
	}

	for _, compatObj := range compatList.Items {
		convertedBuilder, err := b.convertFromV1Builder(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedBuilder)
	}

	return list, nil
}

func (b *builders) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	watchable, err := b.client.Builders(b.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (b *builders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.Builder, err error) {
	v1Result, err := b.client.Builders(b.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := b.convertFromV1Builder(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

//TODO: bump kpack
func (b *builders) convertFromV1Builder(ctx context.Context, v1Builder *v1alpha1.Builder) (*v1alpha2.Builder, error) {
	resultBuilder := &v1alpha2.Builder{}
	err := resultBuilder.ConvertFrom(ctx, v1Builder)
	if err != nil {
		return nil, err
	}
	return resultBuilder, nil
}

func (b *builders) convertToV1Builder(ctx context.Context, v2Builder *v1alpha2.Builder) (*v1alpha1.Builder, error) {
	resultBuilder := &v1alpha1.Builder{}
	err := v2Builder.ConvertTo(ctx, resultBuilder)
	if err != nil {
		return nil, err
	}
	return resultBuilder, nil
}
