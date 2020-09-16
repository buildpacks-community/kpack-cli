// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"io"

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
		Short:        "Patch an existing builder configuration",
		Long:         ` `,
		Example:      `kp builder patch my-builder`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(flags.namespace)
			if err != nil {
				return err
			}

			name := args[0]
			flags.namespace = cs.Namespace

			cb, err := cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return patch(cb, flags, cmd.OutOrStdout(), cs)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVarP(&flags.outputFormat, "output", "", "yaml", "output format. supported formats are: yaml, json")
	return cmd
}

func patch(bldr *v1alpha1.Builder, flags CommandFlags, writer io.Writer, cs k8s.ClientSet) error {
	patchedBldr := bldr.DeepCopy()

	if flags.tag != "" {
		patchedBldr.Spec.Tag = flags.tag
	}

	if flags.stack != "" {
		patchedBldr.Spec.Stack.Name = flags.stack
	}

	if flags.store != "" {
		patchedBldr.Spec.Store.Name = flags.store
	}

	if flags.order != "" {
		orderEntries, err := builder.ReadOrder(flags.order)
		if err != nil {
			return err
		}

		patchedBldr.Spec.Order = orderEntries
	}

	patch, err := k8s.CreatePatch(bldr, patchedBldr)
	if err != nil {
		return err
	}

	if len(patch) == 0 {
		_, err = fmt.Fprintln(writer, "nothing to patch")
		return err
	}

	if flags.dryRun {
		printer, err := commands.NewResourcePrinter(flags.outputFormat)
		if err != nil {
			return err
		}

		return printer.PrintObject(bldr, writer)
	}

	_, err = cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Patch(bldr.Name, types.MergePatchType, patch)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "\"%s\" patched\n", bldr.Name)
	return err
}
