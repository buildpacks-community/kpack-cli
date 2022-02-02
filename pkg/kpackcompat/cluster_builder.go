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

// clusterBuilders implements ClusterBuilderInterface
type clusterBuilders struct {
	client v1alpha1client.KpackV1alpha1Interface
}

func newClusterBuilders(c *kpackV1alpha1CompatClient) *clusterBuilders {
	return &clusterBuilders{
		client: c.v1alpha1KpackClient,
	}
}

func (b *clusterBuilders) Create(ctx context.Context, clusterBuilder *v1alpha2.ClusterBuilder, opts metav1.CreateOptions) (*v1alpha2.ClusterBuilder, error) {
	convertedClusterBuilder, err := convertToV1ClusterBuilder(ctx, clusterBuilder)
	if err != nil {
		return nil, err
	}

	createdV1ClusterBuilder, err := b.client.ClusterBuilders().Create(ctx, convertedClusterBuilder, opts)
	if err != nil {
		return nil, err
	}

	createdV2ClusterBuilder, err := convertFromV1ClusterBuilder(ctx, createdV1ClusterBuilder)
	if err != nil {
		return nil, err
	}

	return createdV2ClusterBuilder, nil
}

func (b *clusterBuilders) Update(ctx context.Context, clusterBuilder *v1alpha2.ClusterBuilder, opts metav1.UpdateOptions) (*v1alpha2.ClusterBuilder, error) {
	convertedClusterBuilder, err := convertToV1ClusterBuilder(ctx, clusterBuilder)
	if err != nil {
		return nil, err
	}

	updatedV1ClusterBuilder, err := b.client.ClusterBuilders().Update(ctx, convertedClusterBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterBuilder, err := convertFromV1ClusterBuilder(ctx, updatedV1ClusterBuilder)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterBuilder, nil
}

func (b *clusterBuilders) UpdateStatus(ctx context.Context, clusterBuilder *v1alpha2.ClusterBuilder, opts metav1.UpdateOptions) (*v1alpha2.ClusterBuilder, error) {
	convertedClusterBuilder, err := convertToV1ClusterBuilder(ctx, clusterBuilder)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := b.client.ClusterBuilders().UpdateStatus(ctx, convertedClusterBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterBuilder, err := convertFromV1ClusterBuilder(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterBuilder, nil
}

func (b *clusterBuilders) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := b.client.ClusterBuilders().Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (b *clusterBuilders) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	err := b.client.ClusterBuilders().DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b *clusterBuilders) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.ClusterBuilder, error) {
	v1ClusterBuilder, err := b.client.ClusterBuilders().Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedClusterBuilder, err := convertFromV1ClusterBuilder(ctx, v1ClusterBuilder)
	if err != nil {
		return nil, err
	}
	return convertedClusterBuilder, nil
}

func (b *clusterBuilders) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.ClusterBuilderList, error) {
	compatList, err := b.client.ClusterBuilders().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.ClusterBuilderList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.ClusterBuilder{},
	}

	for _, compatObj := range compatList.Items {
		convertedClusterBuilder, err := convertFromV1ClusterBuilder(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedClusterBuilder)
	}

	return list, nil
}

func (b *clusterBuilders) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	watchable, err := b.client.ClusterBuilders().Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (b *clusterBuilders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterBuilder, err error) {
	v1Result, err := b.client.ClusterBuilders().Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := convertFromV1ClusterBuilder(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func convertFromV1ClusterBuilder(ctx context.Context, v1ClusterBuilder *v1alpha1.ClusterBuilder) (*v1alpha2.ClusterBuilder, error) {
	resultClusterBuilder := &v1alpha2.ClusterBuilder{}
	err := resultClusterBuilder.ConvertFrom(ctx, v1ClusterBuilder)
	if err != nil {
		return nil, err
	}
	return resultClusterBuilder, nil
}

func convertToV1ClusterBuilder(ctx context.Context, v2ClusterBuilder *v1alpha2.ClusterBuilder) (*v1alpha1.ClusterBuilder, error) {
	resultClusterBuilder := &v1alpha1.ClusterBuilder{}
	err := v2ClusterBuilder.ConvertTo(ctx, resultClusterBuilder)
	if err != nil {
		return nil, err
	}
	return resultClusterBuilder, nil
}
