// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle

import (
	"io"
	"strings"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		verbose bool
	)

	cmd := &cobra.Command{
		Use:          "status <name>",
		Short:        "Display cluster lifecycle status",
		Long:         `Prints detailed information about the status of a specific cluster-scoped lifecycle.`,
		Example:      "kp clusterlifecycle status my-lifecycle",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			lifecycle, err := cs.KpackClient.KpackV1alpha2().ClusterLifecycles().Get(cmd.Context(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return displayLifecycleStatus(cmd.OutOrStdout(), lifecycle, verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "display mixins")

	return cmd
}

func displayLifecycleStatus(out io.Writer, l *v1alpha2.ClusterLifecycle, verbose bool) error {
	writer := commands.NewStatusWriter(out)

	items := []string{
		"Status", getStatusText(l),
		"Image", l.Status.Image.LatestImage,
		"Version", l.Status.Version,
	}

	if verbose {
		items = append(items, "Supported Buildpack APIs", strings.Join(l.Status.APIs.Buildpack.Supported, ", "))
		items = append(items, "Deprecated Buildpack APIs", strings.Join(l.Status.APIs.Buildpack.Deprecated, ", "))
	}

	if err := writer.AddBlock("", items...); err != nil {
		return err
	}

	return writer.Write()
}

func getStatusText(l *v1alpha2.ClusterLifecycle) string {
	if cond := l.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == corev1.ConditionTrue {
			return "Ready"
		} else if cond.Status == corev1.ConditionFalse {
			return "Not Ready - " + cond.Message
		}
	}
	return "Unknown"
}
