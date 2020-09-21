// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:          "patch <name>",
		Short:        "Patch an existing cluster builder configuration",
		Long:         ` `,
		Example:      `kp cb patch my-builder`,
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

			name := args[0]

			cb, err := cs.KpackClient.KpackV1alpha1().ClusterBuilders().Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return patch(cb, flags, ch, cs)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&flags.output, "output", "", "output format. supported formats are: yaml, json")
	return cmd
}

func patch(cb *v1alpha1.ClusterBuilder, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet) error {
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

	if flags.order != "" {
		orderEntries, err := builder.ReadOrder(flags.order)
		if err != nil {
			return err
		}

		patchedCb.Spec.Order = orderEntries
	}

	patch, err := k8s.CreatePatch(cb, patchedCb)
	if err != nil {
		return err
	}

	if len(patch) == 0 {
		return ch.PrintResult("nothing to patch")
	}

	if !ch.IsDryRun() {
		patchedCb, err = cs.KpackClient.KpackV1alpha1().ClusterBuilders().Patch(patchedCb.Name, types.MergePatchType, patch)
		if err != nil {
			return err
		}
	}

	err = ch.PrintObj(patchedCb)
	if err != nil {
		return err
	}

	return ch.PrintResult("%q patched", patchedCb.Name)
}
