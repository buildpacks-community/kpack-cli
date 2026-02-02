// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/builder"
	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/dockercreds"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags     CommandFlags
		tlsConfig registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing cluster builder configuration",
		Long: `Patch an existing clusterbuilder configuration by providing command line arguments.

A buildpack order must be provided with either the path to an order yaml, via the --buildpack flag, or extracted from a builder image using --order-from.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group.`,
		Example: `kp clusterbuilder patch my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp clusterbuilder patch my-builder --order /path/to/order.yaml
kp clusterbuilder patch my-builder --order-from paketobuildpacks/builder-jammy-base
kp clusterbuilder patch my-builder --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1`,
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

			fetcher := registry.NewDefaultFetcher(tlsConfig)
			return patch(ctx, cb, flags, ch, cs, fetcher, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().StringSliceVarP(&flags.buildpacks, "buildpack", "b", []string{}, "buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'\n  repeat for each buildpack in order, or supply once with comma-separated list")
	cmd.Flags().StringVar(&flags.orderFrom, "order-from", "", "builder image to extract buildpack order from")
	commands.SetDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsConfig)
	return cmd
}

func patch(ctx context.Context, cb *v1alpha2.ClusterBuilder, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, fetcher builder.Fetcher, waiter commands.ResourceWaiter) error {
	updatedCb := cb.DeepCopy()

	if flags.tag != "" {
		updatedCb.Spec.Tag = flags.tag
	}

	if flags.stack != "" {
		updatedCb.Spec.Stack.Name = flags.stack
	}

	if flags.store != "" {
		updatedCb.Spec.Store.Name = flags.store
	}

	// Validate that only one order source is provided
	orderSourceCount := 0
	if len(flags.buildpacks) > 0 {
		orderSourceCount++
	}
	if flags.order != "" {
		orderSourceCount++
	}
	if flags.orderFrom != "" {
		orderSourceCount++
	}
	if orderSourceCount > 1 {
		return fmt.Errorf("only one of --order, --buildpack, or --order-from can be specified")
	}

	// Set the order based on the provided flag
	var err error
	if flags.order != "" {
		orderEntries, err := builder.ReadOrder(flags.order)
		if err != nil {
			return err
		}

		updatedCb.Spec.Order = orderEntries
	} else if len(flags.buildpacks) > 0 {
		updatedCb.Spec.Order = builder.CreateOrder(flags.buildpacks)
	} else if flags.orderFrom != "" {
		keychain := dockercreds.DefaultKeychain
		updatedCb.Spec.Order, err = builder.ReadOrderFromImage(keychain, fetcher, flags.orderFrom)
		if err != nil {
			return err
		}
	}

	patch, err := k8s.CreatePatch(cb, updatedCb)
	if err != nil {
		return err
	}

	hasPatch := len(patch) > 0
	if hasPatch && !ch.IsDryRun() {
		updatedCb, err = cs.KpackClient.KpackV1alpha2().ClusterBuilders().Patch(ctx, updatedCb.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		if err := waiter.Wait(ctx, updatedCb); err != nil {
			return err
		}
	}

	updatedCbArray := []runtime.Object{updatedCb}

	if err = ch.PrintObjs(updatedCbArray); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasPatch, "ClusterBuilder %q patched", updatedCb.Name)
}
