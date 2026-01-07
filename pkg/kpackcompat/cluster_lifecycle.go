package kpackcompat

import (
	"context"

	v1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
)

// clusterLifecycles implement ClusterLifecycleInterface
// ClusterLifecycle is only available in kpack v1alpha2 API, so all operations
// return ErrV1alpha2Required when using a v1alpha1-only client.
type clusterLifecycles struct{}

func newClusterLifecycles(c *kpackV1alpha1CompatClient) *clusterLifecycles {
	return &clusterLifecycles{}
}

func (*clusterLifecycles) Create(ctx context.Context, clusterLifecycle *v1alpha2.ClusterLifecycle, opts v1.CreateOptions) (*v1alpha2.ClusterLifecycle, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterLifecycles) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return ErrV1alpha2Required
}

func (*clusterLifecycles) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return ErrV1alpha2Required
}

func (*clusterLifecycles) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.ClusterLifecycle, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterLifecycles) List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.ClusterLifecycleList, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterLifecycles) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterLifecycle, err error) {
	return nil, ErrV1alpha2Required
}

func (*clusterLifecycles) Update(ctx context.Context, clusterLifecycle *v1alpha2.ClusterLifecycle, opts v1.UpdateOptions) (*v1alpha2.ClusterLifecycle, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterLifecycles) UpdateStatus(ctx context.Context, clusterLifecycle *v1alpha2.ClusterLifecycle, opts v1.UpdateOptions) (*v1alpha2.ClusterLifecycle, error) {
	return nil, ErrV1alpha2Required
}

func (*clusterLifecycles) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, ErrV1alpha2Required
}
