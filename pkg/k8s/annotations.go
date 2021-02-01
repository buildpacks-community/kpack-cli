// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

func MergeAnnotations(a1, a2 map[string]string) map[string]string {
	mergedMap := map[string]string{}

	for k, v := range a1 {
		mergedMap[k] = v
	}
	for k, v := range a2 {
		mergedMap[k] = v
	}
	return mergedMap
}
