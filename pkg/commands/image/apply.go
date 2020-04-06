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
		path string
	)

	applyCmd := &ApplyCommand{
		Out:     out,
		Applier: applier,
	}

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply an image configuration",
		Long:    "Apply an image configuration by filename. This image will be created if it doesn't exist yet.\nOnly YAML files are accepted.",
		Example: "tbctl image apply -f ./image.yaml\ncat ./image.yaml | tbctl image apply -f -",
		RunE: func(_ *cobra.Command, args []string) error {
			return applyCmd.Execute(path, args...)
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&path, "file", "f", "", "path to the image configuration file")
	_ = cmd.MarkFlagRequired("file")

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
	var (
		file io.ReadCloser
		err  error
	)

	if path == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(path)
		if err != nil {
			return err
		}
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
