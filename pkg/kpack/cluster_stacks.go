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

// clusterStacks implements ClusterStackInterface
type clusterStacks struct {
	client rest.Interface
	ns     string
}

func (s clusterStacks) Create(ctx context.Context, clusterStack *v1alpha2.ClusterStack, opts metav1.CreateOptions) (*v1alpha2.ClusterStack, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedClusterStack, err := s.convertToV1ClusterStack(ctx, clusterStack)
	if err != nil {
		return nil, err
	}

	createdV1ClusterStack, err := v1Client.ClusterStacks().Create(ctx, convertedClusterStack, opts)
	if err != nil {
		return nil, err
	}

	createdV2ClusterStack, err := s.convertFromV1ClusterStack(ctx, createdV1ClusterStack)
	if err != nil {
		return nil, err
	}

	return createdV2ClusterStack, nil
}

func (s clusterStacks) Update(ctx context.Context, clusterStack *v1alpha2.ClusterStack, opts metav1.UpdateOptions) (*v1alpha2.ClusterStack, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedClusterStack, err := s.convertToV1ClusterStack(ctx, clusterStack)
	if err != nil {
		return nil, err
	}

	updatedV1ClusterStack, err := v1Client.ClusterStacks().Update(ctx, convertedClusterStack, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterStack, err := s.convertFromV1ClusterStack(ctx, updatedV1ClusterStack)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterStack, nil
}

func (s clusterStacks) UpdateStatus(ctx context.Context, clusterStack *v1alpha2.ClusterStack, opts metav1.UpdateOptions) (*v1alpha2.ClusterStack, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedClusterStack, err := s.convertToV1ClusterStack(ctx, clusterStack)
	if err != nil {
		return nil, err
	}

	updatedV1image, err := v1Client.ClusterStacks().UpdateStatus(ctx, convertedClusterStack, opts)
	if err != nil {
		return nil, err
	}

	updatedV2ClusterStack, err := s.convertFromV1ClusterStack(ctx, updatedV1image)
	if err != nil {
		return nil, err
	}

	return updatedV2ClusterStack, nil
}

func (s clusterStacks) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.ClusterStacks().Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (s clusterStacks) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.ClusterStacks().DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (s clusterStacks) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.ClusterStack, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1ClusterStack, err := v1Client.ClusterStacks().Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedClusterStack, err := s.convertFromV1ClusterStack(ctx, v1ClusterStack)
	if err != nil {
		return nil, err
	}
	return convertedClusterStack, nil
}

func (s clusterStacks) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.ClusterStackList, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}
	compatList, err := v1Client.ClusterStacks().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.ClusterStackList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.ClusterStack{},
	}

	for _, compatObj := range compatList.Items {
		convertedClusterStack, err := s.convertFromV1ClusterStack(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedClusterStack)
	}

	return list, nil
}

func (s clusterStacks) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	watchable, err := v1Client.ClusterStacks().Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (s clusterStacks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.ClusterStack, err error) {
	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Result, err := v1Client.ClusterStacks().Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := s.convertFromV1ClusterStack(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

// newClusterStacks returns a ClusterStacks
func newClusterStacks(c *KpackV1alpha2Client) *clusterStacks {
	return &clusterStacks{
		client: c.RESTClient(),
	}
}

//TODO: bump kpack
func (s *clusterStacks) convertFromV1ClusterStack(ctx context.Context, v1ClusterStack *v1alpha1.ClusterStack) (result *buildV1alpha2.ClusterStack, err error) {
	resultClusterStack := buildV1alpha2.ClusterStack{}
	err = resultClusterStack.ConvertFrom(ctx, v1ClusterStack)
	if err != nil {
		return nil, err
	}
	return &resultClusterStack, nil
}

func (s *clusterStacks) convertToV1ClusterStack(ctx context.Context, v2ClusterStack *buildV1alpha2.ClusterStack) (result *v1alpha1.ClusterStack, err error) {
	resultClusterStack := v1alpha1.ClusterStack{}
	err = resultClusterStack.ConvertTo(ctx, v2ClusterStack)
	if err != nil {
		return nil, err
	}
	return &resultClusterStack, nil
}
