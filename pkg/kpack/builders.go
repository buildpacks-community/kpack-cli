package kpack

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	buildV1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1alpha1Client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

// builders implements BuilderInterface
type builders struct {
	client rest.Interface
	ns     string
}

func (b builders) Create(ctx context.Context, builder *v1alpha2.Builder, opts metav1.CreateOptions) (*v1alpha2.Builder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedBuilder, err := b.convertToV1Builder(ctx, builder)
	if err != nil {
		return nil, err
	}

	createdV1Builder, err := v1Client.Builders(b.ns).Create(ctx, convertedBuilder, opts)
	if err != nil {
		return nil, err
	}

	createdV2Builder, err := b.convertFromV1Builder(ctx, createdV1Builder)
	if err != nil {
		return nil, err
	}

	return createdV2Builder, nil
}

func (b builders) Update(ctx context.Context, builder *v1alpha2.Builder, opts metav1.UpdateOptions) (*v1alpha2.Builder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedBuilder, err := b.convertToV1Builder(ctx, builder)
	if err != nil {
		return nil, err
	}

	updatedV1Builder, err := v1Client.Builders(b.ns).Update(ctx, convertedBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Builder, err := b.convertFromV1Builder(ctx, updatedV1Builder)
	if err != nil {
		return nil, err
	}

	return updatedV2Builder, nil
}

func (b builders) UpdateStatus(ctx context.Context, builder *v1alpha2.Builder, opts metav1.UpdateOptions) (*v1alpha2.Builder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedBuilder, err := b.convertToV1Builder(ctx, builder)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := v1Client.Builders(b.ns).UpdateStatus(ctx, convertedBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Builder, err := b.convertFromV1Builder(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2Builder, nil
}

func (b builders) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.Builders(b.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (b builders) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.Builders(b.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b builders) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.Builder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Builder, err := v1Client.Builders(b.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedBuilder, err := b.convertFromV1Builder(ctx, v1Builder)
	if err != nil {
		return nil, err
	}
	return convertedBuilder, nil
}

func (b builders) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.BuilderList, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}
	compatList, err := v1Client.Builders(b.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.BuilderList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.Builder{},
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

func (b builders) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	watchable, err := v1Client.Builders(b.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (b builders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.Builder, err error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Result, err := v1Client.Builders(b.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := b.convertFromV1Builder(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

// newBuilders returns a Builders
func newBuilders(c *KpackV1alpha2Client, namespace string) *builders {
	return &builders{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

//TODO: bump kpack
func (b *builders) convertFromV1Builder(ctx context.Context, v1Builder *v1alpha1.Builder) (result *buildV1alpha2.Builder, err error) {
	resultBuilder := buildV1alpha2.Builder{}
	err = resultBuilder.ConvertFrom(ctx, v1Builder)
	if err != nil {
		return nil, err
	}
	return &resultBuilder, nil
}

func (b *builders) convertToV1Builder(ctx context.Context, v2Builder *buildV1alpha2.Builder) (result *v1alpha1.Builder, err error) {
	resultBuilder := v1alpha1.Builder{}
	err = resultBuilder.ConvertTo(ctx, v2Builder)
	if err != nil {
		return nil, err
	}
	return &resultBuilder, nil
}
