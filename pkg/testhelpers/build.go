// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func MakeTestBuilds(image string, namespace string) []*v1alpha2.Build {
	buildOne := &v1alpha2.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-one",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha2.ImageLabel:       image,
				v1alpha2.BuildNumberLabel: "1",
			},
			Annotations: map[string]string{
				v1alpha2.BuildReasonAnnotation: "CONFIG",
			},
		},
		Spec: v1alpha2.BuildSpec{
			Builder: corev1alpha1.BuildBuilderSpec{
				Image: "some-repo.com/my-builder",
			},
		},
		Status: v1alpha2.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionTrue,
						LastTransitionTime: corev1alpha1.VolatileTime{
							Inner: metav1.Time{},
						},
					},
				},
			},
			BuildMetadata: corev1alpha1.BuildpackMetadataList{
				{
					Id:       "bp-id-1",
					Version:  "bp-version-1",
					Homepage: "mysupercoolsite.com",
				},
				{
					Id:       "bp-id-2",
					Version:  "bp-version-2",
					Homepage: "mysupercoolsite2.com",
				},
			},
			Stack: corev1alpha1.BuildStack{
				RunImage: "some-repo.com/run-image",
			},
			LatestImage: "repo.com/image-1:tag",
			PodName:     "pod-one",
		},
	}
	buildTwo := &v1alpha2.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "build-two",
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Time{}.Add(1 * time.Hour)},
			Labels: map[string]string{
				v1alpha2.ImageLabel:       image,
				v1alpha2.BuildNumberLabel: "2",
			},
			Annotations: map[string]string{
				v1alpha2.BuildReasonAnnotation: "COMMIT,BUILDPACK",
			},
		},
		Status: v1alpha2.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionFalse,
						LastTransitionTime: corev1alpha1.VolatileTime{
							Inner: metav1.Time{},
						},
					},
				},
			},
			LatestImage: "repo.com/image-2:tag",
			PodName:     "pod-two",
		},
	}
	buildThree := &v1alpha2.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "build-three",
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
			Labels: map[string]string{
				v1alpha2.ImageLabel:       image,
				v1alpha2.BuildNumberLabel: "3",
			},
			Annotations: map[string]string{
				v1alpha2.BuildReasonAnnotation: "TRIGGER",
			},
		},
		Spec: v1alpha2.BuildSpec{
			Builder: corev1alpha1.BuildBuilderSpec{
				Image: "some-repo.com/my-builder",
			},
		},
		Status: v1alpha2.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
			BuildMetadata: corev1alpha1.BuildpackMetadataList{
				{
					Id:       "bp-id-1",
					Version:  "bp-version-1",
					Homepage: "mysupercoolsite.com",
				},
				{
					Id:       "bp-id-2",
					Version:  "bp-version-2",
					Homepage: "mysupercoolsite2.com",
				},
			},
			Stack: corev1alpha1.BuildStack{
				RunImage: "some-repo.com/run-image",
			},
			LatestImage: "repo.com/image-3:tag",
			PodName:     "pod-three",
		},
	}
	otherBuild := &v1alpha2.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ignored",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha2.ImageLabel:       "some-other-image",
				v1alpha2.BuildNumberLabel: "1",
			},
		},
		Status: v1alpha2.BuildStatus{
			LatestImage: "repo.com/other-image-1:tag",
		},
	}
	return []*v1alpha2.Build{buildOne, buildThree, buildTwo, otherBuild}
}

func BuildsToRuntimeObjs(builds []*v1alpha2.Build) []runtime.Object {
	var final []runtime.Object
	for _, t := range builds {
		final = append(final, t)
	}
	return final
}
