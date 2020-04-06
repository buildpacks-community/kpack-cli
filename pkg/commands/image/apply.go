package image

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
)

func NewApplyCommand(out io.Writer, applier Applier) *cobra.Command {
	var (
		file string
	)

	applyCmd := &ApplyCommand{
		Out:     out,
		Applier: applier,
	}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply an image configuration",
		RunE: func(_ *cobra.Command, args []string) error {
			return applyCmd.Execute(file, args...)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to the image config")

	return cmd
}

type Applier interface {
	Apply(image *v1alpha1.Image) error
}

type ApplyCommand struct {
	Out     io.Writer
	Applier Applier
}

func (a *ApplyCommand) Execute(path string, args ...string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var imageConfig v1alpha1.Image
	err = yaml.Unmarshal(buf, &imageConfig)
	if err != nil {
		return err
	}

	err = a.Applier.Apply(&imageConfig)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(a.Out, "%s created\n", imageConfig.Name)
	return err
}
