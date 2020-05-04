package custombuilder

import (
	"fmt"
	"io"

	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewStatusCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:          "status <name>",
		Short:        "Display custom builder status",
		Long:         `Prints detailed information about the status of a specific custom builder.`,
		Example:      "tbctl cb status my-builder\ntbctl cb status -n my-namespace other-builder",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			bldr, err := kpackClient.ExperimentalV1alpha1().CustomBuilders(namespace).Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return displayBuilderStatus(bldr, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return cmd
}

func displayBuilderStatus(bldr *expv1alpha1.CustomBuilder, writer io.Writer) error {
	if cond := bldr.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == v1.ConditionTrue {
			return printBuilderReadyStatus(bldr, writer)
		} else {
			return printBuilderNotReadyStatus(bldr, writer)
		}
	} else {
		return printBuilderConditionUnknownStatus(bldr, writer)
	}
}

func printBuilderConditionUnknownStatus(_ *expv1alpha1.CustomBuilder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	return statusWriter.AddBlock(
		"",
		"Status", "Unknown",
	)
}

func printBuilderNotReadyStatus(bldr *expv1alpha1.CustomBuilder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	condReady := bldr.Status.GetCondition(corev1alpha1.ConditionReady)

	return statusWriter.AddBlock(
		"",
		"Status", "Not Ready",
		"Reason", condReady.Message,
	)
}

func printBuilderReadyStatus(bldr *expv1alpha1.CustomBuilder, writer io.Writer) error {
	statusWriter := commands.NewStatusWriter(writer)

	err := statusWriter.AddBlock(
		"",
		"Status", "Ready",
		"Image", bldr.Status.LatestImage,
		"Stack", bldr.Status.Stack.ID,
		"Run Image", bldr.Status.Stack.RunImage,
	)

	if err != nil {
		return err
	}

	bpTableWriter, err := commands.NewTableWriter(writer, "buildpack id", "version")
	if err != nil {
		return nil
	}

	for _, bpMD := range bldr.Status.BuilderMetadata {
		err := bpTableWriter.AddRow(bpMD.Id, bpMD.Version)
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

	for i, entry := range bldr.Spec.Order {
		err := orderTableWriter.AddRow(fmt.Sprintf("Group #%d", i+1), "")
		if err != nil {
			return err
		}
		for _, ref := range entry.Group {
			if ref.Optional {
				err := orderTableWriter.AddRow("  "+ref.Id, "(Optional)")
				if err != nil {
					return err
				}
			} else {
				err := orderTableWriter.AddRow("  "+ref.Id, "")
				if err != nil {
					return err
				}
			}
		}
	}
	return orderTableWriter.Write()
}
