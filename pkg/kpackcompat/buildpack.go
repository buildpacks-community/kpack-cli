package kpackcompat

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// buildpacks implements BuildpackInterface
type buildpacks struct{}

func newBuildpacks(c *kpackV1alpha1CompatClient, namespace string) *buildpacks {
	return &buildpacks{}
}

func (*buildpacks) Create(ctx context.Context, buildpack *v1alpha2.Buildpack, opts v1.CreateOptions) (*v1alpha2.Buildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*buildpacks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return ErrV1alpha2Required
}

func (*buildpacks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return ErrV1alpha2Required
}

func (*buildpacks) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.Buildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*buildpacks) List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.BuildpackList, error) {
	return nil, ErrV1alpha2Required
}

func (*buildpacks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Buildpack, err error) {
	return nil, ErrV1alpha2Required
}

func (*buildpacks) Update(ctx context.Context, buildpack *v1alpha2.Buildpack, opts v1.UpdateOptions) (*v1alpha2.Buildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*buildpacks) UpdateStatus(ctx context.Context, buildpack *v1alpha2.Buildpack, opts v1.UpdateOptions) (*v1alpha2.Buildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*buildpacks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, ErrV1alpha2Required
}
