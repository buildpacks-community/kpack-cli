package kpack

import (
	"context"

	buildV1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *kpackClient) ListBuilds(ctx context.Context, namespace string, opts v1.ListOptions) (*buildV1alpha2.BuildList, error) {
	if k.groupVersion == kpackGroupVersionV1alpha2 {
		return k.client.KpackV1alpha2().Builds(namespace).List(ctx, opts)
	}

	compatList, err := k.client.KpackV1alpha1().Builds(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.BuildList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.Build{},
	}

	for _, compatObj := range compatList.Items {
		list.Items = append(list.Items, buildV1alpha2.Build{
			TypeMeta:   compatObj.TypeMeta,
			ObjectMeta: compatObj.ObjectMeta,
			Spec: buildV1alpha2.BuildSpec{
				Tags:                  compatObj.Spec.Tags,
				Builder:               compatObj.Spec.Builder,
				ServiceAccount:        compatObj.Spec.ServiceAccount,
				Source:                compatObj.Spec.Source,
				Cache:                 nil,
				Bindings:              compatObj.Spec.Bindings,
				Env:                   compatObj.Spec.Env,
				ProjectDescriptorPath: "",
				Resources:             compatObj.Spec.Resources,
				LastBuild:             nil,
				Notary:                compatObj.Spec.Notary,
			},
			Status: buildV1alpha2.BuildStatus{
				Status:           compatObj.Status.Status,
				BuildMetadata:    compatObj.Status.BuildMetadata,
				Stack:            compatObj.Status.Stack,
				LatestImage:      compatObj.Status.LatestImage,
				LatestCacheImage: "",
				PodName:          compatObj.Status.PodName,
				StepStates:       compatObj.Status.StepStates,
				StepsCompleted:   compatObj.Status.StepsCompleted,
			},
		})
	}
	return list, nil
}
