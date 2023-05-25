package kpackcompat

import (
	"context"

	v1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
)

// clusterBuildpacks implement ClusterBuildpackInterface
type clusterBuildpacks struct{}

func newClusterBuildpacks(c *kpackV1alpha1CompatClient) *clusterBuildpacks {
	return &clusterBuildpacks{}
}

func (*clusterBuildpacks) Create(ctx context.Context, clusterBuildpack *v1alpha2.ClusterBuildpack, opts v1.CreateOptions) (*v1alpha2.ClusterBuildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterBuildpacks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return ErrV1alpha2Required
}

func (*clusterBuildpacks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return ErrV1alpha2Required
}

func (*clusterBuildpacks) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.ClusterBuildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterBuildpacks) List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.ClusterBuildpackList, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterBuildpacks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterBuildpack, err error) {
	return nil, ErrV1alpha2Required
}

func (*clusterBuildpacks) Update(ctx context.Context, clusterBuildpack *v1alpha2.ClusterBuildpack, opts v1.UpdateOptions) (*v1alpha2.ClusterBuildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterBuildpacks) UpdateStatus(ctx context.Context, clusterBuildpack *v1alpha2.ClusterBuildpack, opts v1.UpdateOptions) (*v1alpha2.ClusterBuildpack, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterBuildpacks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, ErrV1alpha2Required
}
