// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"

func Sort(builds []v1alpha1.Build) func(i int, j int) bool {
	return func(i, j int) bool {
		l1, _ := builds[i].ObjectMeta.Labels[v1alpha1.ImageLabel]
		l2, _ := builds[j].ObjectMeta.Labels[v1alpha1.ImageLabel]
		if l1 != l2 {
			return l1 > l2
		}

		return builds[j].ObjectMeta.CreationTimestamp.After(builds[i].ObjectMeta.CreationTimestamp.Time)
	}
}
