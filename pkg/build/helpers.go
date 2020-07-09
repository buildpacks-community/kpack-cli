// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

func Sort(builds []v1alpha1.Build) func(i int, j int) bool {
	return func(i, j int) bool {
		return builds[j].ObjectMeta.CreationTimestamp.After(builds[i].ObjectMeta.CreationTimestamp.Time)
	}
}
