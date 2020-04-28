package stack

import (
	"io"
	"strings"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewStatusCommand(kpackClient versioned.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status <name>",
		Short:        "Display stack status",
		Long:         `Prints detailed information about the status of a stack.`,
		Example:      "tbctl stack status my-stack",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := kpackClient.ExperimentalV1alpha1().Stacks().Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return displayStackStatus(cmd.OutOrStdout(), s)
		},
	}
	return cmd
}

func displayStackStatus(out io.Writer, s *expv1alpha1.Stack) error {
	writer := commands.NewStatusWriter(out)

	err := writer.AddBlock("",
		"Id", s.Status.Id,
		"Run Image", s.Status.RunImage.LatestImage,
		"Build Image", s.Status.BuildImage.LatestImage,
		"Mixins", strings.Join(s.Status.Mixins, ", "),
	)
	if err != nil {
		return err
	}

	return writer.Write()
}
