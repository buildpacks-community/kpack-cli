// Copyright 2020-2020 VMware, Inc.
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
	cmd := &cobra.Command{
		Use:   "remove <store> <buildpackage> [<buildpackage>...]",
		Short: "Remove buildpackage(s) from cluster store",
		Long: `Removes existing buildpackage(s) from a specific cluster-scoped buildpack store.

This relies on the image(s) specified to exist in the store and removes the associated buildpackage(s)
`,
		Example: `kp clusterstore remove my-store my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271
kp clusterstore remove my-store my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271 my-registry.com/my-buildpackage/buildpacks_nginx@sha256:eacecf914a57e4f5d6c4459226f6133ecaccb5c462717a09cfeae4763207b9ef
`,
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			printer := commands.NewPrinter(cmd)

			storeName, buildPackages := args[0], args[1:]

			store, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(storeName, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("ClusterStore '%s' does not exist", storeName)
			} else if err != nil {
				return err
			}

			for _, bpToRemove := range buildPackages {
				if !storeContainsBuildpackage(store, bpToRemove) {
					return errors.Errorf("Buildpackage '%s' does not exist in the clusterstore", bpToRemove)
				}
			}
			var updatedStoreSources = []v1alpha1.StoreImage{}
			for _, storeImg := range store.Spec.Sources {
				found := false
				for _, bpToRemove := range args {
					if storeImg.Image == bpToRemove {
						found = true
						printer.Printf("Removing buildpackage %s", bpToRemove)
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

			printer.Printf("ClusterStore Updated")
			return nil
		},
	}
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
