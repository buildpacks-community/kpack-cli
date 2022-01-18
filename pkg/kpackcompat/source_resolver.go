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

// sourceResolvers implements SourceResolverInterface
type sourceResolvers struct {
	client v1alpha1client.KpackV1alpha1Interface
	ns     string
}

// newSourceResolvers returns a SourceResolvers
func newSourceResolvers(c *KpackV1alpha1CompatClient, namespace string) *sourceResolvers {
	return &sourceResolvers{
		client: c.v1alpha1KpackClient,
		ns:     namespace,
	}
}

func (s *sourceResolvers) Create(ctx context.Context, sourceResolver *v1alpha2.SourceResolver, opts metav1.CreateOptions) (*v1alpha2.SourceResolver, error) {
	convertedSourceResolver, err := s.convertToV1SourceResolver(ctx, sourceResolver)
	if err != nil {
		return nil, err
	}

	createdV1SourceResolver, err := s.client.SourceResolvers(s.ns).Create(ctx, convertedSourceResolver, opts)
	if err != nil {
		return nil, err
	}

	createdV2SourceResolver, err := s.convertFromV1SourceResolver(ctx, createdV1SourceResolver)
	if err != nil {
		return nil, err
	}

	return createdV2SourceResolver, nil
}

func (s *sourceResolvers) Update(ctx context.Context, sourceResolver *v1alpha2.SourceResolver, opts metav1.UpdateOptions) (*v1alpha2.SourceResolver, error) {
	convertedSourceResolver, err := s.convertToV1SourceResolver(ctx, sourceResolver)
	if err != nil {
		return nil, err
	}

	updatedV1sourceResolver, err := s.client.SourceResolvers(s.ns).Update(ctx, convertedSourceResolver, opts)
	if err != nil {
		return nil, err
	}

	updatedV2SourceResolver, err := s.convertFromV1SourceResolver(ctx, updatedV1sourceResolver)
	if err != nil {
		return nil, err
	}

	return updatedV2SourceResolver, nil
}

func (s *sourceResolvers) UpdateStatus(ctx context.Context, sourceResolver *v1alpha2.SourceResolver, opts metav1.UpdateOptions) (*v1alpha2.SourceResolver, error) {
	convertedSourceResolver, err := s.convertToV1SourceResolver(ctx, sourceResolver)
	if err != nil {
		return nil, err
	}

	updatedV1sourceResolver, err := s.client.SourceResolvers(s.ns).UpdateStatus(ctx, convertedSourceResolver, opts)
	if err != nil {
		return nil, err
	}

	updatedV2SourceResolver, err := s.convertFromV1SourceResolver(ctx, updatedV1sourceResolver)
	if err != nil {
		return nil, err
	}

	return updatedV2SourceResolver, nil
}

func (s *sourceResolvers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := s.client.SourceResolvers(s.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (s *sourceResolvers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	err := s.client.SourceResolvers(s.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (s *sourceResolvers) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha2.SourceResolver, error) {
	v1SourceResolver, err := s.client.SourceResolvers(s.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedSourceResolver, err := s.convertFromV1SourceResolver(ctx, v1SourceResolver)
	if err != nil {
		return nil, err
	}
	return convertedSourceResolver, nil
}

func (s *sourceResolvers) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha2.SourceResolverList, error) {
	compatList, err := s.client.SourceResolvers(s.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &v1alpha2.SourceResolverList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []v1alpha2.SourceResolver{},
	}

	for _, compatObj := range compatList.Items {
		convertedSourceResolver, err := s.convertFromV1SourceResolver(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedSourceResolver)
	}

	return list, nil
}

func (s *sourceResolvers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	watchable, err := s.client.SourceResolvers(s.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (s *sourceResolvers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha2.SourceResolver, err error) {
	v1Result, err := s.client.SourceResolvers(s.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := s.convertFromV1SourceResolver(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func (s *sourceResolvers) convertFromV1SourceResolver(ctx context.Context, v1SourceResolver *v1alpha1.SourceResolver) (*v1alpha2.SourceResolver, error) {
	resultSourceResolver := &v1alpha2.SourceResolver{}
	err := resultSourceResolver.ConvertFrom(ctx, v1SourceResolver)
	if err != nil {
		return nil, err
	}
	return resultSourceResolver, nil
}

func (s *sourceResolvers) convertToV1SourceResolver(ctx context.Context, v2SourceResolver *v1alpha2.SourceResolver) (*v1alpha1.SourceResolver, error) {
	resultSourceResolver := &v1alpha1.SourceResolver{}
	err := v2SourceResolver.ConvertTo(ctx, resultSourceResolver)
	if err != nil {
		return nil, err
	}
	return resultSourceResolver, nil
}
