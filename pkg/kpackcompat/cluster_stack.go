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

// clusterStacks implements ClusterStackInterface
type clusterStacks struct {
	client v1alpha1client.KpackV1alpha1Interface
}

func newClusterStacks(c *kpackV1alpha1CompatClient) *clusterStacks {
	return &clusterStacks{
		client: c.v1alpha1KpackClient,
	}
}

func (s *clusterStacks) Create(ctx context.Context, clusterStack *v1alpha2.ClusterStack, opts metav1.CreateOptions) (*v1alpha2.ClusterStack, error) {
	convertedClusterStack, err := convertToV1ClusterStack(ctx, clusterStack)
	if err != nil {
		return nil, err
	}

	createdV1ClusterStack, err := s.client.ClusterStacks().Create(ctx, convertedClusterStack, opts)
	if err != nil {
		return nil, err
	}

	createdV2ClusterStack, err := convertFromV1ClusterStack(ctx, createdV1ClusterStack)
	if err != nil {
		return nil, err
	}

	return createdV2ClusterStack, nil
}

func (s *clusterStacks) Update(ctx context.Context, clusterStack *v1alpha2.ClusterStack, opts metav1.UpdateOptions) (*v1alpha2.ClusterStack, error) {
	convertedClusterStack, err := convertToV1ClusterStack(ctx, clusterStack)
	if err != nil {
		return nil, err
	}

	updatedV1ClusterStack, err := s.client.ClusterStacks().Update(ctx, convertedClusterStack, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterStack, err := convertFromV1ClusterStack(ctx, updatedV1ClusterStack)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterStack, nil
}

func (s *clusterStacks) UpdateStatus(ctx context.Context, clusterStack *v1alpha2.ClusterStack, opts metav1.UpdateOptions) (*v1alpha2.ClusterStack, error) {
	convertedClusterStack, err := convertToV1ClusterStack(ctx, clusterStack)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := s.client.ClusterStacks().UpdateStatus(ctx, convertedClusterStack, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterStack, err := convertFromV1ClusterStack(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterStack, nil
}

func (s *clusterStacks) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := s.client.ClusterStacks().Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (s *clusterStacks) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	err := s.client.ClusterStacks().DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (s *clusterStacks) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.ClusterStack, error) {
	v1ClusterStack, err := s.client.ClusterStacks().Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedClusterStack, err := convertFromV1ClusterStack(ctx, v1ClusterStack)
	if err != nil {
		return nil, err
	}
	return convertedClusterStack, nil
}

func (s *clusterStacks) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.ClusterStackList, error) {
	compatList, err := s.client.ClusterStacks().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.ClusterStackList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.ClusterStack{},
	}

	for _, compatObj := range compatList.Items {
		convertedClusterStack, err := convertFromV1ClusterStack(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedClusterStack)
	}

	return list, nil
}

func (s *clusterStacks) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	watchable, err := s.client.ClusterStacks().Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (s *clusterStacks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterStack, err error) {
	v1Result, err := s.client.ClusterStacks().Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := convertFromV1ClusterStack(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func convertFromV1ClusterStack(ctx context.Context, v1ClusterStack *v1alpha1.ClusterStack) (*v1alpha2.ClusterStack, error) {
	resultClusterStack := &v1alpha2.ClusterStack{}
	err := resultClusterStack.ConvertFrom(ctx, v1ClusterStack)
	if err != nil {
		return nil, err
	}
	return resultClusterStack, nil
}

func convertToV1ClusterStack(ctx context.Context, v2ClusterStack *v1alpha2.ClusterStack) (*v1alpha1.ClusterStack, error) {
	resultClusterStack := &v1alpha1.ClusterStack{}
	err := v2ClusterStack.ConvertTo(ctx, resultClusterStack)
	if err != nil {
		return nil, err
	}
	return resultClusterStack, nil
}
