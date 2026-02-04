// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package descriptor

const APIVersionV1 = "kp.kpack.io/v1"

// ClusterLifecycle represents a ClusterLifecycle in the v1 descriptor
type ClusterLifecycle struct {
	Name  string `yaml:"name" json:"name"`
	Image string `yaml:"image" json:"image"`
}

// ClusterBuildpack represents a ClusterBuildpack in the v1 descriptor
type ClusterBuildpack struct {
	Name  string `yaml:"name" json:"name"`
	Image string `yaml:"image" json:"image"`
}

// DependencyDescriptor represents the target v1 format that all conversions produce
type DependencyDescriptor struct {
	APIVersion              string             `yaml:"apiVersion" json:"apiVersion"`
	Kind                    string             `yaml:"kind" json:"kind"`
	DefaultClusterLifecycle string             `yaml:"defaultClusterLifecycle,omitempty" json:"defaultClusterLifecycle,omitempty"`
	DefaultClusterStack     string             `yaml:"defaultClusterStack,omitempty" json:"defaultClusterStack,omitempty"`
	DefaultClusterBuilder   string             `yaml:"defaultClusterBuilder,omitempty" json:"defaultClusterBuilder,omitempty"`
	ClusterLifecycles       []ClusterLifecycle `yaml:"clusterLifecycles,omitempty" json:"clusterLifecycles,omitempty"`
	ClusterBuildpacks       []ClusterBuildpack `yaml:"clusterBuildpacks,omitempty" json:"clusterBuildpacks,omitempty"`
	ClusterStores           []ClusterStore     `yaml:"clusterStores,omitempty" json:"clusterStores,omitempty"`
	ClusterStacks           []ClusterStack     `yaml:"clusterStacks,omitempty" json:"clusterStacks,omitempty"`
	ClusterBuilders         []ClusterBuilder   `yaml:"clusterBuilders,omitempty" json:"clusterBuilders,omitempty"`
}
