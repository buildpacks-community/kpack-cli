package kpack

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	buildV1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1alpha1Client "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

// builds implements BuildInterface
type builds struct {
	client rest.Interface
	ns     string
}

// newBuilds returns a Builds
func newBuilds(c *KpackV1alpha2Client, namespace string) *builds {
	return &builds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *builds) Create(ctx context.Context, build *buildV1alpha2.Build, opts v1.CreateOptions) (*buildV1alpha2.Build, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedBuild, err := c.convertToV1Build(ctx, build)
	if err != nil {
		return nil, err
	}

	createdV1Build, err := v1Client.Builds(c.ns).Create(ctx, convertedBuild, opts)
	if err != nil {
		return nil, err
	}

	createdV2Build, err := c.convertFromV1Build(ctx, createdV1Build)
	if err != nil {
		return nil, err
	}

	return createdV2Build, nil
}

func (c *builds) Update(ctx context.Context, build *buildV1alpha2.Build, opts v1.UpdateOptions) (*buildV1alpha2.Build, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedBuild, err := c.convertToV1Build(ctx, build)
	if err != nil {
		return nil, err
	}

	updatedV1build, err := v1Client.Builds(c.ns).Update(ctx, convertedBuild, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Build, err := c.convertFromV1Build(ctx, updatedV1build)
	if err != nil {
		return nil, err
	}

	return updatedV2Build, nil
}

func (c *builds) UpdateStatus(ctx context.Context, build *buildV1alpha2.Build, opts v1.UpdateOptions) (*buildV1alpha2.Build, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	convertedBuild, err := c.convertToV1Build(ctx, build)
	if err != nil {
		return nil, err
	}

	updatedV1build, err := v1Client.Builds(c.ns).UpdateStatus(ctx, convertedBuild, opts)
	if err != nil {
		return nil, err
	}

	updatedV2Build, err := c.convertFromV1Build(ctx, updatedV1build)
	if err != nil {
		return nil, err
	}

	return updatedV2Build, nil
}

func (c *builds) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.Builds(c.ns).Delete(ctx, name, opts)
	if err != nil {
		return err
	}
	return nil
}

func (c *builds) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	err := v1Client.Builds(c.ns).DeleteCollection(ctx, opts, listOpts)
	if err != nil {
		return err
	}
	return nil
}

func (c *builds) Get(ctx context.Context, name string, opts v1.GetOptions) (*buildV1alpha2.Build, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Build, err := v1Client.Builds(c.ns).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}

	convertedBuild, err := c.convertFromV1Build(ctx, v1Build)
	if err != nil {
		return nil, err
	}
	return convertedBuild, nil
}

func (c *builds) List(ctx context.Context, opts v1.ListOptions) (*buildV1alpha2.BuildList, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}
	compatList, err := v1Client.Builds(c.ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.BuildList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.Build{},
	}

	for _, compatObj := range compatList.Items {
		convertedBuild, err := c.convertFromV1Build(ctx, &compatObj)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertedBuild)
	}

	return list, nil
}

func (c *builds) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	watchable, err := v1Client.Builds(c.ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watchable, nil
}

func (c *builds) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *buildV1alpha2.Build, err error) {

	v1Client := v1alpha1Client.KpackV1alpha1Client{}

	v1Result, err := v1Client.Builds(c.ns).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}

	v2Result, err := c.convertFromV1Build(ctx, v1Result)
	if err != nil {
		return nil, err
	}

	return v2Result, nil
}

func (c *builds) convertFromV1Build(ctx context.Context, build *v1alpha1.Build) (result *buildV1alpha2.Build, err error) {
	resultBuild := buildV1alpha2.Build{}
	err = resultBuild.ConvertFrom(ctx, build)
	if err != nil {
		return nil, err
	}
	return &resultBuild, nil
}

func (c *builds) convertToV1Build(ctx context.Context, v2Build *buildV1alpha2.Build) (result *v1alpha1.Build, err error) {
	resultBuild := v1alpha1.Build{}
	err = resultBuild.ConvertTo(ctx, v2Build)
	if err != nil {
		return nil, err
	}
	return &resultBuild, nil
}
