package kpackcompat

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1alpha1client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// builds implements BuildInterface
type builds struct {
	client v1alpha1client.KpackV1alpha1Interface
	ns     string
}

func newBuilds(c *kpackV1alpha1CompatClient, namespace string) *builds {
	return &builds{
		client: c.v1alpha1KpackClient,
		ns:     namespace,
	}
}

func (b *builds) Create(ctx context.Context, build *v1alpha2.Build, opts v1.CreateOptions) (*v1alpha2.Build, error) {
	convertedBuild, err := convertToV1Build(ctx, build)
	if err != nil {
		return nil, err
	}

	createdV1Build, err := b.client.Builds(b.ns).Create(ctx, convertedBuild, opts)
	if err != nil {
		return nil, err
	}

	createdV2Build, err := convertFromV1Build(ctx, createdV1Build)
	if err != nil {
		return nil, err
	}

	return createdV2Build, nil
}

func (b *builds) Update(ctx context.Context, build *v1alpha2.Build, opts v1.UpdateOptions) (*v1alpha2.Build, error) {
	convertedBuild, err := convertToV1Build(ctx, build)
	if err != nil {
		return nil, err
	}

	updatedV1build, err := b.client.Builds(b.ns).Update(ctx, convertedBuild, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Build, err := convertFromV1Build(ctx, updatedV1build)
	if err != nil {
		return nil, err
	}

	return updatedV2Build, nil
}

func (b *builds) UpdateStatus(ctx context.Context, build *v1alpha2.Build, opts v1.UpdateOptions) (*v1alpha2.Build, error) {
	convertedBuild, err := convertToV1Build(ctx, build)
	if err != nil {
		return nil, err
	}

	updatedV1build, err := b.client.Builds(b.ns).UpdateStatus(ctx, convertedBuild, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Build, err := convertFromV1Build(ctx, updatedV1build)
	if err != nil {
		return nil, err
	}

	return updatedV2Build, nil
}

func (b *builds) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	err := b.client.Builds(b.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (b *builds) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	err := b.client.Builds(b.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b *builds) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.Build, error) {
	v1Build, err := b.client.Builds(b.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedBuild, err := convertFromV1Build(ctx, v1Build)
	if err != nil {
		return nil, err
	}
	return convertedBuild, nil
}

func (b *builds) List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.BuildList, error) {
	compatList, err := b.client.Builds(b.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.BuildList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.Build{},
	}

	for _, compatObj := range compatList.Items {
		convertedBuild, err := convertFromV1Build(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedBuild)
	}

	return list, nil
}

func (b *builds) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	watchable, err := b.client.Builds(b.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (b *builds) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Build, err error) {
	v1Result, err := b.client.Builds(b.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := convertFromV1Build(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func convertFromV1Build(ctx context.Context, build *v1alpha1.Build) (*v1alpha2.Build, error) {
	resultBuild := &v1alpha2.Build{}
	err := resultBuild.ConvertFrom(ctx, build)
	if err != nil {
		return nil, err
	}
	return resultBuild, nil
}

func convertToV1Build(ctx context.Context, v2Build *v1alpha2.Build) (*v1alpha1.Build, error) {
	resultBuild := &v1alpha1.Build{}
	err := v2Build.ConvertTo(ctx, resultBuild)
	if err != nil {
		return nil, err
	}
	return resultBuild, nil
}
