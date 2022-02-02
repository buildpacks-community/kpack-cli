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

// clusterStores implements ClusterStoreInterface
type clusterStores struct {
	client v1alpha1client.KpackV1alpha1Interface
}

func newClusterStores(c *kpackV1alpha1CompatClient) *clusterStores {
	return &clusterStores{
		client: c.v1alpha1KpackClient,
	}
}

func (s *clusterStores) Create(ctx context.Context, clusterStore *v1alpha2.ClusterStore, opts metav1.CreateOptions) (*v1alpha2.ClusterStore, error) {
	convertedClusterStore, err := s.convertToV1ClusterStore(ctx, clusterStore)
	if err != nil {
		return nil, err
	}

	createdV1ClusterStore, err := s.client.ClusterStores().Create(ctx, convertedClusterStore, opts)
	if err != nil {
		return nil, err
	}

	createdV2ClusterStore, err := s.convertFromV1ClusterStore(ctx, createdV1ClusterStore)
	if err != nil {
		return nil, err
	}

	return createdV2ClusterStore, nil
}

func (s *clusterStores) Update(ctx context.Context, clusterStore *v1alpha2.ClusterStore, opts metav1.UpdateOptions) (*v1alpha2.ClusterStore, error) {
	convertedClusterStore, err := s.convertToV1ClusterStore(ctx, clusterStore)
	if err != nil {
		return nil, err
	}

	updatedV1ClusterStore, err := s.client.ClusterStores().Update(ctx, convertedClusterStore, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterStore, err := s.convertFromV1ClusterStore(ctx, updatedV1ClusterStore)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterStore, nil
}

func (s *clusterStores) UpdateStatus(ctx context.Context, clusterStore *v1alpha2.ClusterStore, opts metav1.UpdateOptions) (*v1alpha2.ClusterStore, error) {
	convertedClusterStore, err := s.convertToV1ClusterStore(ctx, clusterStore)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := s.client.ClusterStores().UpdateStatus(ctx, convertedClusterStore, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterStore, err := s.convertFromV1ClusterStore(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterStore, nil
}

func (s *clusterStores) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := s.client.ClusterStores().Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (s *clusterStores) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	err := s.client.ClusterStores().DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (s *clusterStores) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.ClusterStore, error) {
	v1ClusterStore, err := s.client.ClusterStores().Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedClusterStore, err := s.convertFromV1ClusterStore(ctx, v1ClusterStore)
	if err != nil {
		return nil, err
	}
	return convertedClusterStore, nil
}

func (s *clusterStores) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.ClusterStoreList, error) {
	compatList, err := s.client.ClusterStores().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.ClusterStoreList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.ClusterStore{},
	}

	for _, compatObj := range compatList.Items {
		convertedClusterStore, err := s.convertFromV1ClusterStore(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedClusterStore)
	}

	return list, nil
}

func (s *clusterStores) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	watchable, err := s.client.ClusterStores().Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (s *clusterStores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterStore, err error) {
	v1Result, err := s.client.ClusterStores().Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := s.convertFromV1ClusterStore(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func (s *clusterStores) convertFromV1ClusterStore(ctx context.Context, v1ClusterStore *v1alpha1.ClusterStore) (*v1alpha2.ClusterStore, error) {
	resultClusterStore := &v1alpha2.ClusterStore{}
	err := resultClusterStore.ConvertFrom(ctx, v1ClusterStore)
	if err != nil {
		return nil, err
	}
	return resultClusterStore, nil
}

func (s *clusterStores) convertToV1ClusterStore(ctx context.Context, v2ClusterStore *v1alpha2.ClusterStore) (*v1alpha1.ClusterStore, error) {
	resultClusterStore := &v1alpha1.ClusterStore{}
	err := v2ClusterStore.ConvertTo(ctx, resultClusterStore)
	if err != nil {
		return nil, err
	}
	return resultClusterStore, nil
}
