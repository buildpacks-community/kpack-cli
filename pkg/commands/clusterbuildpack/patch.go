// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack

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
		Use:          "patch <name>",
		Short:        "Patch an existing cluster buildpack configuration",
		Long:         "Patch an existing cluster buildpack configuration by providing command line arguments.",
		Example:      "kp cbp patch my-buildpack --image gcr.io/paketo-buildpacks/java",
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
			cbp, err := cs.KpackClient.KpackV1alpha2().ClusterBuildpacks().Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return patch(ctx, cbp, flags, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "registry location where the buildpack is located")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}

func patch(ctx context.Context, cbp *v1alpha2.ClusterBuildpack, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) error {
	updatedCbp := cbp.DeepCopy()

	if flags.image != "" {
		updatedCbp.Spec.Image = flags.image
	}

	patch, err := k8s.CreatePatch(cbp, updatedCbp)
	if err != nil {
		return err
	}

	hasPatch := len(patch) > 0
	if hasPatch && !ch.IsDryRun() {
		updatedCbp, err = cs.KpackClient.KpackV1alpha2().ClusterBuildpacks().Patch(ctx, updatedCbp.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		if err = w.Wait(ctx, updatedCbp); err != nil {
			return err
		}
	}

	if err = ch.PrintObj(updatedCbp); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasPatch, "Cluster Buildpack %q patched", updatedCbp.Name)
}
