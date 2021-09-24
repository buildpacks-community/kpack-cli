package kpack

import (
	"context"

	buildV1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *kpackClient) ListImages(ctx context.Context, namespace string, opts v1.ListOptions) (*buildV1alpha2.ImageList, error) {
	if k.groupVersion == kpackGroupVersionV1alpha2 {
		return k.client.KpackV1alpha2().Images(namespace).List(ctx, opts)
	}

	compatList, err := k.client.KpackV1alpha1().Images(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	list := &buildV1alpha2.ImageList{
		TypeMeta: compatList.TypeMeta,
		ListMeta: compatList.ListMeta,
		Items:    []buildV1alpha2.Image{},
	}

	for _, compatObj := range compatList.Items {
		list.Items = append(list.Items, buildV1alpha2.Image{
			TypeMeta:   compatObj.TypeMeta,
			ObjectMeta: compatObj.ObjectMeta,
			Spec: buildV1alpha2.ImageSpec{
				Tag:                      compatObj.Spec.Tag,
				Builder:                  compatObj.Spec.Builder,
				ServiceAccount:           compatObj.Spec.ServiceAccount,
				Source:                   compatObj.Spec.Source,
				Cache:                    nil,
				FailedBuildHistoryLimit:  compatObj.Spec.FailedBuildHistoryLimit,
				SuccessBuildHistoryLimit: compatObj.Spec.SuccessBuildHistoryLimit,
				ImageTaggingStrategy:     compatObj.Spec.ImageTaggingStrategy,
				ProjectDescriptorPath:    "",
				Build:                    compatObj.Spec.Build,
				Notary:                   compatObj.Spec.Notary,
			},
			Status: buildV1alpha2.ImageStatus{
				Status:                     compatObj.Status.Status,
				LatestBuildRef:             compatObj.Status.LatestBuildRef,
				LatestBuildImageGeneration: compatObj.Status.LatestBuildImageGeneration,
				LatestImage:                compatObj.Status.LatestImage,
				LatestStack:                compatObj.Status.LatestStack,
				BuildCounter:               compatObj.Status.BuildCounter,
				BuildCacheName:             compatObj.Status.BuildCacheName,
				LatestBuildReason:          compatObj.Status.LatestBuildReason,
			},
		})
	}

	return list, nil
}
