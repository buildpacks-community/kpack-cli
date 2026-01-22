// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package descriptor

import (
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"

	"github.com/buildpacks-community/kpack-cli/pkg/builder"
)

const APIVersionV1Alpha1 = "kp.kpack.io/v1alpha1"

type ClusterBuilderV1Alpha1 struct {
	Name  string                    `yaml:"name"`
	Stack string                    `yaml:"stack"`
	Store string                    `yaml:"store"`
	Order []corev1alpha1.OrderEntry `yaml:"order"`
}

// DependencyDescriptorV1Alpha1 represents the v1alpha1 format of the dependency descriptor
type DependencyDescriptorV1Alpha1 struct {
	APIVersion            string                   `yaml:"apiVersion"`
	Kind                  string                   `yaml:"kind"`
	DefaultStack          string                   `yaml:"defaultStack"`
	DefaultClusterBuilder string                   `yaml:"defaultClusterBuilder"`
	Stores                []ClusterStore           `yaml:"stores"`
	Stacks                []ClusterStack           `yaml:"stacks"`
	ClusterBuilders       []ClusterBuilderV1Alpha1 `yaml:"clusterBuilders"`
}

// ToV1 converts a v1alpha1 descriptor to the v1 format
func (d1 DependencyDescriptorV1Alpha1) ToV1() DependencyDescriptor {
	var d DependencyDescriptor
	d.APIVersion = APIVersionV1
	d.Kind = d1.Kind
	d.DefaultClusterStack = d1.DefaultStack
	d.DefaultClusterBuilder = d1.DefaultClusterBuilder
	d.ClusterStores = d1.Stores
	d.ClusterStacks = d1.Stacks

	// v1alpha1 doesn't have lifecycle or buildpacks, so these will be empty
	d.ClusterLifecycles = []ClusterLifecycle{}
	d.ClusterBuildpacks = []ClusterBuildpack{}
	for _, cb := range d1.ClusterBuilders {
		d.ClusterBuilders = append(d.ClusterBuilders, ClusterBuilder{
			Name:         cb.Name,
			ClusterStack: cb.Stack,
			ClusterStore: cb.Store,
			Order:        builder.CoreOrderEntryToBuildOrderEntry(cb.Order),
		})
	}
	return d
}
