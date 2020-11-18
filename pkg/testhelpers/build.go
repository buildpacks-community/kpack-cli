// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func MakeTestBuilds(image string, namespace string) []runtime.Object {
	buildOne := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-one",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.ImageLabel:       image,
				v1alpha1.BuildNumberLabel: "1",
			},
			Annotations: map[string]string{
				v1alpha1.BuildReasonAnnotation: "CONFIG",
			},
		},
		Spec: v1alpha1.BuildSpec{
			Builder: v1alpha1.BuildBuilderSpec{
				Image: "some-repo.com/my-builder",
			},
		},
		Status: v1alpha1.BuildStatus{
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
			BuildMetadata: v1alpha1.BuildpackMetadataList{
				{
					Id:      "bp-id-1",
					Version: "bp-version-1",
				},
				{
					Id:      "bp-id-2",
					Version: "bp-version-2",
				},
			},
			Stack: v1alpha1.BuildStack{
				RunImage: "some-repo.com/run-image",
			},
			LatestImage: "repo.com/image-1:tag",
			PodName:     "pod-one",
		},
	}
	buildTwo := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "build-two",
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Time{}.Add(1 * time.Hour)},
			Labels: map[string]string{
				v1alpha1.ImageLabel:       image,
				v1alpha1.BuildNumberLabel: "2",
			},
			Annotations: map[string]string{
				v1alpha1.BuildReasonAnnotation: "COMMIT,BUILDPACK",
			},
		},
		Status: v1alpha1.BuildStatus{
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
	buildThree := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "build-three",
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
			Labels: map[string]string{
				v1alpha1.ImageLabel:       image,
				v1alpha1.BuildNumberLabel: "3",
			},
			Annotations: map[string]string{
				v1alpha1.BuildReasonAnnotation: "TRIGGER",
			},
		},
		Spec: v1alpha1.BuildSpec{
			Builder: v1alpha1.BuildBuilderSpec{
				Image: "some-repo.com/my-builder",
			},
		},
		Status: v1alpha1.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
			BuildMetadata: v1alpha1.BuildpackMetadataList{
				{
					Id:      "bp-id-1",
					Version: "bp-version-1",
				},
				{
					Id:      "bp-id-2",
					Version: "bp-version-2",
				},
			},
			Stack: v1alpha1.BuildStack{
				RunImage: "some-repo.com/run-image",
			},
			LatestImage: "repo.com/image-3:tag",
			PodName:     "pod-three",
		},
	}
	otherBuild := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ignored",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.ImageLabel:       "some-other-image",
				v1alpha1.BuildNumberLabel: "1",
			},
		},
		Status: v1alpha1.BuildStatus{
			LatestImage: "repo.com/other-image-1:tag",
		},
	}
	return []runtime.Object{buildOne, buildThree, buildTwo, otherBuild}
}
