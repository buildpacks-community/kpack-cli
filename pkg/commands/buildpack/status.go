// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack

import (
	"io"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Display status of a buildpack",
		Long: `Prints detailed information about the status of a specific buildpack in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp buildpack status my-buildpack\nkp buildpack status -n my-namespace other-buildpack",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			bp, err := cs.KpackClient.KpackV1alpha2().Buildpacks(cs.Namespace).Get(cmd.Context(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return displayBuildpackStatus(bp, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}

func displayBuildpackStatus(bp *v1alpha2.Buildpack, writer io.Writer) error {
	if cond := bp.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == corev1.ConditionTrue {
			return printBuildpackReadyStatus(bp, writer)
		} else {
			return printBuildpackNotReadyStatus(bp, writer)
		}
	} else {
		return printBuildpackConditionUnknownStatus(bp, writer)
	}
}

func printBuildpackConditionUnknownStatus(_ *v1alpha2.Buildpack, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	return statusWriter.AddBlock(
		"",
		"Status", "Unknown",
	)
}

func printBuildpackNotReadyStatus(bp *v1alpha2.Buildpack, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	condReady := bp.Status.GetCondition(corev1alpha1.ConditionReady)

	return statusWriter.AddBlock(
		"",
		"Status", "Not Ready",
		"Reason", condReady.Message,
	)
}

func printBuildpackReadyStatus(bp *v1alpha2.Buildpack, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	err := statusWriter.AddBlock(
		"",
		"Status", "Ready",
		"Source", bp.Spec.Image,
	)
	if err != nil {
		return err
	}

	bpTableWriter, err := commands.NewTableWriter(writer, "buildpack id", "version", "homepage")
	if err != nil {
		return nil
	}

	for _, bpStatus := range bp.Status.Buildpacks {
		err = bpTableWriter.AddRow(bpStatus.Id, bpStatus.Version, bpStatus.Homepage)
		if err != nil {
			return err
		}
	}

	return bpTableWriter.Write()
}
