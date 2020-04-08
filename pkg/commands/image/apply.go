package image

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/image"
)

func NewApplyCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		path string
	)

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply an image configuration",
		Long:    "Apply an image configuration by filename. This image will be created if it doesn't exist yet.\nOnly YAML files are accepted.",
		Example: "tbctl image apply -f ./image.yaml\ncat ./image.yaml | tbctl image apply -f -",
		RunE: func(cmd *cobra.Command, args []string) error {

			applyCmd := &ApplyCommand{
				Out:              cmd.OutOrStdout(),
				Applier:          &image.Applier{KpackClient: kpackClient},
				DefaultNamespace: defaultNamespace,
			}

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
	Out              io.Writer
	Applier          Applier
	DefaultNamespace string
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

	if imageConfig.Namespace == "" {
		imageConfig.Namespace = a.DefaultNamespace
	}

	err = a.Applier.Apply(&imageConfig)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(a.Out, "%s created\n", imageConfig.Name)
	return err
}
