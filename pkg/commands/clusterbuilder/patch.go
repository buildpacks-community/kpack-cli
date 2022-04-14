// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"context"
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/builder"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing cluster builder configuration",
		Long: `Patch an existing clusterbuilder configuration by providing command line arguments.

A buildpack order must be provided with either the path to an order yaml or via the --buildpack flag.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group.`,
		Example: `kp cb patch my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb patch my-builder --order /path/to/order.yaml
kp cb patch my-builder --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1`,
		Args:         commands.ExactArgsWithUsage(1),
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

			name := args[0]

			ctx := cmd.Context()
			cb, err := cs.KpackClient.KpackV1alpha2().ClusterBuilders().Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return patch(ctx, cb, flags, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().StringSliceVarP(&flags.buildpacks, "buildpack", "b", []string{}, "buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'\n  repeat for each buildpack in order, or supply once with comma-separated list")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}

func patch(ctx context.Context, cb *v1alpha2.ClusterBuilder, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, waiter commands.ResourceWaiter) error {
	patchedCb := cb.DeepCopy()

	if flags.tag != "" {
		patchedCb.Spec.Tag = flags.tag
	}

	if flags.stack != "" {
		patchedCb.Spec.Stack.Name = flags.stack
	}

	if flags.store != "" {
		patchedCb.Spec.Store.Name = flags.store
	}

	if len(flags.buildpacks) > 0 && flags.order != "" {
		return fmt.Errorf("cannot use --order and --buildpack together")
	}

	if flags.order != "" {
		orderEntries, err := builder.ReadOrder(flags.order)
		if err != nil {
			return err
		}

		patchedCb.Spec.Order = orderEntries
	}

	if len(flags.buildpacks) > 0 {
		patchedCb.Spec.Order = builder.CreateOrder(flags.buildpacks)
	}

	patch, err := k8s.CreatePatch(cb, patchedCb)
	if err != nil {
		return err
	}

	hasPatch := len(patch) > 0
	if hasPatch && !ch.IsDryRun() {
		patchedCb, err = cs.KpackClient.KpackV1alpha2().ClusterBuilders().Patch(ctx, patchedCb.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		if err := waiter.Wait(ctx, patchedCb); err != nil {
			return err
		}
	}

	if err = ch.PrintObj(patchedCb); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasPatch, "ClusterBuilder %q patched", patchedCb.Name)
}
