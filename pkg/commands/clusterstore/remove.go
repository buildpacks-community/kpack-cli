// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewRemoveCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var buildpackages []string

	cmd := &cobra.Command{
		Use:   "remove <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Remove buildpackage(s) from cluster store",
		Long: `Removes existing buildpackage(s) from a specific cluster-scoped buildpack store.

This relies on the image(s) specified to exist in the store and removes the associated buildpackage(s)
`,
		Example: `kp clusterstore remove my-store -b my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271
kp clusterstore remove my-store -b my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271 -b my-registry.com/my-buildpackage/buildpacks_nginx@sha256:eacecf914a57e4f5d6c4459226f6133ecaccb5c462717a09cfeae4763207b9ef
`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			storeName := args[0]

			store, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(storeName, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("ClusterStore '%s' does not exist", storeName)
			} else if err != nil {
				return err
			}

			for _, bpToRemove := range buildpackages {
				if !storeContainsBuildpackage(store, bpToRemove) {
					return errors.Errorf("Buildpackage '%s' does not exist in the clusterstore", bpToRemove)
				}
			}
			var updatedStoreSources []v1alpha1.StoreImage
			for _, storeImg := range store.Spec.Sources {
				found := false
				for _, bpToRemove := range buildpackages {
					if storeImg.Image == bpToRemove {
						found = true
						ch.Printlnf("Removing buildpackage %s", bpToRemove)
						break
					}
				}
				if !found {
					updatedStoreSources = append(updatedStoreSources, storeImg)
				}
			}

			store.Spec.Sources = updatedStoreSources

			_, err = cs.KpackClient.KpackV1alpha1().ClusterStores().Update(store)
			if err != nil {
				return err
			}

			ch.Printlnf("ClusterStore Updated")
			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "buildpackage to remove")

	return cmd
}

func storeContainsBuildpackage(store *v1alpha1.ClusterStore, buildpackage string) bool {
	for _, source := range store.Spec.Sources {
		if source.Image == buildpackage {
			return true
		}
	}
	return false
}
