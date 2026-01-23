// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package descriptor

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
)

// APIVersionV1Alpha3 is the API version string for v1alpha3 descriptors
const APIVersionV1Alpha3 = "kp.kpack.io/v1alpha3"

// Lifecycle represents a single lifecycle image in v1alpha3 format
type Lifecycle struct {
	Image string `yaml:"image" json:"image"`
}

// DependencyDescriptorV1Alpha3 represents the v1alpha3 format of the dependency descriptor
type DependencyDescriptorV1Alpha3 struct {
	APIVersion            string           `yaml:"apiVersion"`
	Kind                  string           `yaml:"kind"`
	DefaultClusterStack   string           `yaml:"defaultClusterStack"`
	DefaultClusterBuilder string           `yaml:"defaultClusterBuilder"`
	Lifecycle             Lifecycle        `yaml:"lifecycle"`
	ClusterStores         []ClusterStore   `yaml:"clusterStores"`
	ClusterStacks         []ClusterStack   `yaml:"clusterStacks"`
	ClusterBuilders       []ClusterBuilder `yaml:"clusterBuilders"`
}

// ToV1 converts a v1alpha3 descriptor to the v1 format
func (d DependencyDescriptorV1Alpha3) ToV1() DependencyDescriptor {
	var v1 DependencyDescriptor
	v1.APIVersion = APIVersionV1
	v1.Kind = d.Kind
	v1.DefaultClusterStack = d.DefaultClusterStack
	v1.DefaultClusterBuilder = d.DefaultClusterBuilder
	v1.ClusterStores = d.ClusterStores
	v1.ClusterStacks = d.ClusterStacks
	v1.ClusterBuilders = d.ClusterBuilders

	// v1alpha3 doesn't have buildpacks
	v1.ClusterBuildpacks = []ClusterBuildpack{}

	// Convert single lifecycle to array with kpack's default lifecycle name
	if d.Lifecycle.Image != "" {
		v1.ClusterLifecycles = []ClusterLifecycle{
			{
				Name:  v1alpha2.DefaultLifecycleName,
				Image: d.Lifecycle.Image,
			},
		}
	} else {
		v1.ClusterLifecycles = []ClusterLifecycle{}
	}

	return v1
}
