// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing buildpack configuration",
		Long: `Patch an existing buildpack configuration by providing command line arguments.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp buildpack patch my-buildpack --image gcr.io/paketo-buildpacks/java
kp buildpack patch my-buildpack --service-account my-other-sa`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(flags.namespace)
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			flags.namespace = cs.Namespace

			ctx := cmd.Context()

			bp, err := cs.KpackClient.KpackV1alpha2().Buildpacks(cs.Namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return patch(ctx, bp, flags, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "registry location where the buildpack is located")
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&flags.serviceAccount, "service-account", "", "service account name to use")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}

func patch(ctx context.Context, bp *v1alpha2.Buildpack, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) error {
	updatedBp := bp.DeepCopy()

	if flags.image != "" {
		updatedBp.Spec.Image = flags.image
	}

	if flags.serviceAccount != "" {
		updatedBp.Spec.ServiceAccountName = flags.serviceAccount
	}

	patch, err := k8s.CreatePatch(bp, updatedBp)
	if err != nil {
		return err
	}

	hasPatch := len(patch) > 0
	if hasPatch && !ch.IsDryRun() {
		updatedBp, err = cs.KpackClient.KpackV1alpha2().Buildpacks(cs.Namespace).Patch(ctx, updatedBp.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		if err = w.Wait(ctx, updatedBp); err != nil {
			return err
		}
	}

	if err = ch.PrintObj(updatedBp); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasPatch, "Buildpack %q patched", updatedBp.Name)
}
