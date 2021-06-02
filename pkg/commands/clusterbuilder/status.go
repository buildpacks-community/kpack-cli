// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"fmt"
	"io"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/builder"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {

	cmd := &cobra.Command{
		Use:          "status <name>",
		Short:        "Display cluster builder status",
		Long:         `Prints detailed information about the status of a specific cluster builder.`,
		Example:      "kp cb status my-builder",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			bldr, err := cs.KpackClient.KpackV1alpha1().ClusterBuilders().Get(cmd.Context(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return displayBuilderStatus(bldr, cmd.OutOrStdout())
		},
	}

	return cmd
}

func displayBuilderStatus(bldr *v1alpha1.ClusterBuilder, writer io.Writer) error {
	if cond := bldr.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == corev1.ConditionTrue {
			return printBuilderReadyStatus(bldr, writer)
		} else {
			return printBuilderNotReadyStatus(bldr, writer)
		}
	} else {
		return printBuilderConditionUnknownStatus(bldr, writer)
	}
}

func printBuilderConditionUnknownStatus(_ *v1alpha1.ClusterBuilder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	return statusWriter.AddBlock(
		"",
		"Status", "Unknown",
	)
}

func printBuilderNotReadyStatus(bldr *v1alpha1.ClusterBuilder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	condReady := bldr.Status.GetCondition(corev1alpha1.ConditionReady)

	return statusWriter.AddBlock(
		"",
		"Status", "Not Ready",
		"Reason", condReady.Message,
	)
}

func printBuilderReadyStatus(bldr *v1alpha1.ClusterBuilder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	err := statusWriter.AddBlock(
		"",
		"Status", "Ready",
		"Image", bldr.Status.LatestImage,
		"Stack ID", bldr.Status.Stack.ID,
		"Run Image", bldr.Status.Stack.RunImage,
	)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"",
		"Stack Ref", " ",
		"  Name", bldr.Spec.Stack.Name,
		"  Kind", bldr.Spec.Stack.Kind,
		"Store Ref", " ",
		"  Name", bldr.Spec.Store.Name,
		"  Kind", bldr.Spec.Store.Kind,
	)
	if err != nil {
		return err
	}

	bpTableWriter, err := commands.NewTableWriter(writer, "buildpack id", "version", "homepage")
	if err != nil {
		return nil
	}

	for _, bpMD := range bldr.Status.BuilderMetadata {
		err := bpTableWriter.AddRow(bpMD.Id, bpMD.Version, bpMD.Homepage)
		if err != nil {
			return err
		}
	}

	err = bpTableWriter.Write()
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte("\n"))
	if err != nil {
		return err
	}

	orderTableWriter, err := commands.NewTableWriter(writer, "Detection Order", "")
	if err != nil {
		return nil
	}

	order := bldr.Status.Order
	if len(order) == 0 {
		order = bldr.Spec.Order
	}

	for i, entry := range order {
		err := orderTableWriter.AddRow(fmt.Sprintf("Group #%d", i+1), "")
		if err != nil {
			return err
		}
		for _, ref := range entry.Group {
			err := orderTableWriter.AddRow(builder.CreateDetectionOrderRow(ref))
			if err != nil {
				return err
			}
		}
	}
	return orderTableWriter.Write()
}
