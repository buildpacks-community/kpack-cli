package image

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/image"
)

func NewCreateCommand(kpackClient versioned.Interface, factory *image.Factory, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "create <name> <tag>",
		Short: "Create an image configuration",
		Long: `Create an image configuration by providing command line arguments.
This image will be created if it does not yet exist.`,
		Example:      "",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			img, err := factory.MakeImage(args[0], namespace, args[1])
			if err != nil {
				return err
			}

			_, err = kpackClient.BuildV1alpha1().Images(namespace).Create(img)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", img.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")
	cmd.Flags().StringVarP(&factory.GitRepo, "git", "", "", "")
	cmd.Flags().StringVarP(&factory.GitRevision, "git-revision", "", "master", "")
	cmd.Flags().StringVarP(&factory.Blob, "blob", "", "", "")
	cmd.Flags().StringVarP(&factory.LocalPath, "local-path", "", "", "")
	cmd.Flags().StringVarP(&factory.SubPath, "sub-path", "", "", "")
	cmd.Flags().StringVarP(&factory.Builder, "builder", "", "", "")
	cmd.Flags().StringVarP(&factory.ClusterBuilder, "cluster-builder", "", "", "")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "", []string{}, "")

	return cmd
}
