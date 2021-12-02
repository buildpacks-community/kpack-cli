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

// clusterBuilders implements ClusterBuilderInterface
type clusterBuilders struct {
	client rest.Interface
	ns     string
}

func (b clusterBuilders) Create(ctx context.Context, clusterBuilder *v1alpha2.ClusterBuilder, opts metav1.CreateOptions) (*v1alpha2.ClusterBuilder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedClusterBuilder, err := b.convertToV1ClusterBuilder(ctx, clusterBuilder)
	if err != nil {
		return nil, err
	}

	createdV1ClusterBuilder, err := v1Client.ClusterBuilders().Create(ctx, convertedClusterBuilder, opts)
	if err != nil {
		return nil, err
	}

	createdV2ClusterBuilder, err := b.convertFromV1ClusterBuilder(ctx, createdV1ClusterBuilder)
	if err != nil {
		return nil, err
	}

	return createdV2ClusterBuilder, nil
}

func (b clusterBuilders) Update(ctx context.Context, clusterBuilder *v1alpha2.ClusterBuilder, opts metav1.UpdateOptions) (*v1alpha2.ClusterBuilder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedClusterBuilder, err := b.convertToV1ClusterBuilder(ctx, clusterBuilder)
	if err != nil {
		return nil, err
	}

	updatedV1ClusterBuilder, err := v1Client.ClusterBuilders().Update(ctx, convertedClusterBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterBuilder, err := b.convertFromV1ClusterBuilder(ctx, updatedV1ClusterBuilder)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterBuilder, nil
}

func (b clusterBuilders) UpdateStatus(ctx context.Context, clusterBuilder *v1alpha2.ClusterBuilder, opts metav1.UpdateOptions) (*v1alpha2.ClusterBuilder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedClusterBuilder, err := b.convertToV1ClusterBuilder(ctx, clusterBuilder)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := v1Client.ClusterBuilders().UpdateStatus(ctx, convertedClusterBuilder, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterBuilder, err := b.convertFromV1ClusterBuilder(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterBuilder, nil
}

func (b clusterBuilders) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.ClusterBuilders().Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (b clusterBuilders) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.ClusterBuilders().DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (b clusterBuilders) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.ClusterBuilder, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1ClusterBuilder, err := v1Client.ClusterBuilders().Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedClusterBuilder, err := b.convertFromV1ClusterBuilder(ctx, v1ClusterBuilder)
	if err != nil {
		return nil, err
	}
	return convertedClusterBuilder, nil
}

func (b clusterBuilders) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.ClusterBuilderList, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}
	compatList, err := v1Client.ClusterBuilders().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.ClusterBuilderList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.ClusterBuilder{},
	}

	for _, compatObj := range compatList.Items {
		convertedClusterBuilder, err := b.convertFromV1ClusterBuilder(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedClusterBuilder)
	}

	return list, nil
}

func (b clusterBuilders) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	watchable, err := v1Client.ClusterBuilders().Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (b clusterBuilders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterBuilder, err error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Result, err := v1Client.ClusterBuilders().Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := b.convertFromV1ClusterBuilder(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

// newClusterBuilders returns a ClusterBuilders
func newClusterBuilders(c *KpackV1alpha2Client) *clusterBuilders {
	return &clusterBuilders{
		client: c.RESTClient(),
	}
}

//TODO: bump kpack
func (b *clusterBuilders) convertFromV1ClusterBuilder(ctx context.Context, v1ClusterBuilder *v1alpha1.ClusterBuilder) (result *buildV1alpha2.ClusterBuilder, err error) {
	resultClusterBuilder := buildV1alpha2.ClusterBuilder{}
	err = resultClusterBuilder.ConvertFrom(ctx, v1ClusterBuilder)
	if err != nil {
		return nil, err
	}
	return &resultClusterBuilder, nil
}

func (b *clusterBuilders) convertToV1ClusterBuilder(ctx context.Context, v2ClusterBuilder *buildV1alpha2.ClusterBuilder) (result *v1alpha1.ClusterBuilder, err error) {
	resultClusterBuilder := v1alpha1.ClusterBuilder{}
	err = resultClusterBuilder.ConvertTo(ctx, v2ClusterBuilder)
	if err != nil {
		return nil, err
	}
	return &resultClusterBuilder, nil
}
