package image

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
)

func NewListCommand(out io.Writer, lister Lister) *cobra.Command {
	var (
		namespace string
	)

	listCmd := &ListCommand{
		Out:    out,
		Lister: lister,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Display images in the desired namespace",
		Long:    "Prints a table of the most important information about images. You will only see images in your current namespace.\nIf no namespace is provided, the default namespace is queried.",
		Example: "tbctl image list\ntbctl image list -n my-namespace",
		RunE: func(_ *cobra.Command, args []string) error {
			return listCmd.Execute(namespace)
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "the namespace to query")

	return cmd
}

type Lister interface {
	List(namespace string) (*v1alpha1.ImageList, error)
}

type ListCommand struct {
	Out    io.Writer
	Lister Lister
}

func (a *ListCommand) Execute(namespace string) error {
	imageList, err := a.Lister.List(namespace)
	if err != nil {
		return err
	}

	if len(imageList.Items) == 0 {
		_, err := fmt.Fprintln(a.Out, "no images found in "+namespace+" namespace")
		return err
	}

	writer := tabwriter.NewWriter(a.Out, 0, 0, 0, ' ', 0)
	_, err = fmt.Fprintln(writer, "Name\tReady\tLatest Image")
	if err != nil {
		return err
	}

	for _, img := range imageList.Items {
		_, err = fmt.Fprintf(writer, "%s\t%s\t%s\n", img.Name, "", img.Status.LatestImage)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
