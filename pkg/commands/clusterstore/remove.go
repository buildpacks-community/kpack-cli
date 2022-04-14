// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/kpack-cli/pkg/clusterstore"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewRemoveCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var buildpackages []string

	cmd := &cobra.Command{
		Use:   "remove <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Remove buildpackage(s) from cluster store",
		Long: `Removes existing buildpackage(s) from a specific cluster-scoped buildpack store.
`,
		Example: `kp clusterstore remove my-store -b buildpackage@1.0.0
kp clusterstore remove my-store -b buildpackage@1.0.0 -b other-buildpackage@2.0.0
`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			w := newWaiter(cs.DynamicClient)

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			storeName := args[0]

			store, err := cs.KpackClient.KpackV1alpha2().ClusterStores().Get(ctx, storeName, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("ClusterStore '%s' does not exist", storeName)
			} else if err != nil {
				return err
			}

			factory := clusterstore.NewFactory(ch, nil, nil)

			if err = ch.PrintStatus("Removing Buildpackages..."); err != nil {
				return err
			}

			updatedStore, err := factory.RemoveFromStore(store, buildpackages...)
			if err != nil {
				return err
			}

			if !ch.IsDryRun() {
				patch, err := k8s.CreatePatch(store, updatedStore)
				if err != nil {
					return err
				}
				updatedStore, err = cs.KpackClient.KpackV1alpha2().ClusterStores().Patch(ctx, updatedStore.Name, types.MergePatchType, patch, metav1.PatchOptions{})
				if err != nil {
					return err
				}
				if err := w.Wait(ctx, updatedStore); err != nil {
					return err
				}
			}

			if err = ch.PrintObj(updatedStore); err != nil {
				return err
			}

			return ch.PrintResult("ClusterStore %q updated", updatedStore.Name)
		},
	}
	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "buildpackage to remove")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}
