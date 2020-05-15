package custombuilder

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewApplyCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		path string
	)

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply a custom builder configuration",
		Long:    "Apply a custom builder configuration by filename.\nThe custom builder will be created if it does not yet exist.\nOnly YAML files are accepted.",
		Example: "tbctl cb apply -f ./builder.yaml\ncat ./builder.yaml | tbctl cb apply -f -",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			builderConfig, err := getBuilderConfig(path)
			if err != nil {
				return err
			}

			if builderConfig.Namespace == "" {
				builderConfig.Namespace = cs.Namespace
			}

			_, err = cs.KpackClient.ExperimentalV1alpha1().CustomBuilders(builderConfig.Namespace).Get(builderConfig.Name, metav1.GetOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				return err
			} else if k8serrors.IsNotFound(err) {
				_, err = cs.KpackClient.ExperimentalV1alpha1().CustomBuilders(builderConfig.Namespace).Create(builderConfig)
			} else {
				_, err = cs.KpackClient.ExperimentalV1alpha1().CustomBuilders(builderConfig.Namespace).Update(builderConfig)
			}
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" applied\n", builderConfig.Name)
			return err
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&path, "file", "f", "", "path to the builder configuration file")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func getBuilderConfig(path string) (*expv1alpha1.CustomBuilder, error) {
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

	var builderConfig expv1alpha1.CustomBuilder
	err = yaml.Unmarshal(buf, &builderConfig)
	if err != nil {
		return nil, err
	}
	return &builderConfig, nil
}
