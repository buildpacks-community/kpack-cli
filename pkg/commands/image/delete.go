package image

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func NewDeleteCommand(out io.Writer, defaultNamespace string, deleter Deleter) *cobra.Command {
	var (
		namespace string
	)

	deleteCmd := &DeleteCommand{
		Out:              out,
		Deleter:          deleter,
		DefaultNamespace: defaultNamespace,
	}

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete an image",
		Long:    "Delete an image and the associated image builds from the cluster.",
		Example: "tbctl image delete my-image",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return deleteCmd.Execute(namespace, args...)
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "the namespace of the image to delete")

	return cmd
}

type Deleter interface {
	Delete(namespace, name string) error
}

type DeleteCommand struct {
	Out              io.Writer
	Deleter          Deleter
	DefaultNamespace string
}

func (d *DeleteCommand) Execute(namespace string, args ...string) error {
	if namespace == "" {
		namespace = d.DefaultNamespace
	}

	err := d.Deleter.Delete(namespace, args[0])
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(d.Out, "%s deleted\n", args[0])
	return err
}
