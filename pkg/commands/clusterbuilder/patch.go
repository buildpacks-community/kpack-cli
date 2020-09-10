// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		stack string
		store string
		order string
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

			name := args[0]

			cb, err := cs.KpackClient.KpackV1alpha1().ClusterBuilders().Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return patch(cb, stack, store, order, cmd, cs)
		},
	}
	cmd.Flags().StringVarP(&stack, "stack", "s", "", "stack resource to use")
	cmd.Flags().StringVar(&store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&order, "order", "o", "", "path to buildpack order yaml")

	return cmd
}

func patch(cb *v1alpha1.ClusterBuilder, stack string, store string, order string, cmd *cobra.Command, cs k8s.ClientSet) error {
	patchedCb := cb.DeepCopy()

	if stack != "" {
		patchedCb.Spec.Stack.Name = stack
	}

	if store != "" {
		patchedCb.Spec.Store.Name = store
	}

	if order != "" {
		orderEntries, err := builder.ReadOrder(order)
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
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "nothing to patch")
		return err
	}

	_, err = cs.KpackClient.KpackV1alpha1().ClusterBuilders().Patch(cb.Name, types.MergePatchType, patch)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" patched\n", cb.Name)
	return err
}
