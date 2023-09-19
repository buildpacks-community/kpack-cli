// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"io"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/builder"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Display status of a builder",
		Long: `Prints detailed information about the status of a specific builder in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp builder status my-builder\nkp builder status -n my-namespace other-builder",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			bldr, err := cs.KpackClient.KpackV1alpha2().Builders(cs.Namespace).Get(cmd.Context(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return displayBuilderStatus(bldr, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}

func displayBuilderStatus(bldr *v1alpha2.Builder, writer io.Writer) error {
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

func printBuilderConditionUnknownStatus(_ *v1alpha2.Builder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	return statusWriter.AddBlock(
		"",
		"Status", "Unknown",
	)
}

func printBuilderNotReadyStatus(bldr *v1alpha2.Builder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	condReady := bldr.Status.GetCondition(corev1alpha1.ConditionReady)

	return statusWriter.AddBlock(
		"",
		"Status", "Not Ready",
		"Reason", condReady.Message,
	)
}

func printBuilderReadyStatus(bldr *v1alpha2.Builder, writer io.Writer) error {
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

	cpTableWriter, err := commands.NewTableWriter(writer, "BuildpackName", "     BuildpackKind")
	  if err != nil {
	  	return nil
	  }

	for _, entry := range bldr.Spec.Order {
		for _, ref := range entry.Group {
			if ref.ObjectReference.Name != "" && ref.ObjectReference.Kind != "" {
			err := cpTableWriter.AddRow(ref.ObjectReference.Name, ref.ObjectReference.Kind)
		if err != nil {
			return err
		}
	}
	 }
	 cpTableWriter.Write()
	}

	orderTableWriter, err := commands.NewTableWriter(writer, "Detection Order", "")
	if err != nil {
		return nil
	}

	for i, entry := range bldr.Status.Order {
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
