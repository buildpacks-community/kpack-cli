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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewApplyCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		path string
	)

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply an image configuration",
		Long:    "Apply an image configuration by filename. The image will be created if it does not yet exist.\nOnly YAML files are accepted.",
		Example: "tbctl image apply -f ./image.yaml\ncat ./image.yaml | tbctl image apply -f -",
		RunE: func(cmd *cobra.Command, args []string) error {
			imageConfig, err := getImageConfig(path)
			if err != nil {
				return err
			}

			if imageConfig.Namespace == "" {
				imageConfig.Namespace = defaultNamespace
			}

			_, err = kpackClient.BuildV1alpha1().Images(imageConfig.Namespace).Get(imageConfig.Name, metav1.GetOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				return err
			} else if k8serrors.IsNotFound(err) {
				_, err = kpackClient.BuildV1alpha1().Images(imageConfig.Namespace).Create(imageConfig)
			} else {
				_, err = kpackClient.BuildV1alpha1().Images(imageConfig.Namespace).Update(imageConfig)
			}
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" applied\n", imageConfig.Name)
			return err
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&path, "file", "f", "", "path to the image configuration file")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func getImageConfig(path string) (*v1alpha1.Image, error) {
	var (
		file io.ReadCloser
		err  error
	)

	if path == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var imageConfig v1alpha1.Image
	err = yaml.Unmarshal(buf, &imageConfig)
	if err != nil {
		return nil, err
	}
	return &imageConfig, nil
}
