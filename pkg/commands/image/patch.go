package image

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/build-service-cli/pkg/image"
)

func NewPatchCommand(kpackClient versioned.Interface, factory *image.PatchFactory, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing image configuration",
		Long: `Patch an existing image configuration by providing command line arguments.
This will fail if the image does not already exist.

The flags for this command determine how the build will retrieve source code:

	"--git" and "--git-revision" to use Git based source

	"--blob" to use source code hosted in a blob store

	"--local-path" to use source code from the local machine

Local source code will be pushed to the same registry as the existing image tag.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by
the key value pair. For example, "--env key1=value1 --env key2=value2 ...".
`,
		Example: `tbctl image patch my-image --git-revision my-other-branch
tbctl image patch my-image --blob https://my-blob-host.com/my-blob
tbctl image patch my-image --local-path /path/to/local/source/code
tbctl image patch my-image --local-path /path/to/local/source/code --builder my-builder -n my-namespace
tbctl image patch my-image --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --delete-env food --delete-env PWD`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			img, err := kpackClient.BuildV1alpha1().Images(namespace).Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			patch, err := factory.MakePatch(img)
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), string(patch))
			return nil

			_, err = kpackClient.BuildV1alpha1().Images(namespace).Patch(args[0], types.JSONPatchType, patch)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", img.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")
	cmd.Flags().StringVarP(&factory.GitRepo, "git", "", "", "git repository url")
	cmd.Flags().StringVarP(&factory.GitRevision, "git-revision", "", "master", "git revision")
	cmd.Flags().StringVarP(&factory.Blob, "blob", "", "", "source code blob url")
	cmd.Flags().StringVarP(&factory.LocalPath, "local-path", "", "", "path to local source code")
	cmd.Flags().StringVarP(&factory.SubPath, "sub-path", "", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVarP(&factory.Builder, "builder", "", "", "builder name")
	cmd.Flags().StringVarP(&factory.ClusterBuilder, "cluster-builder", "", "", "cluster builder name")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "", []string{}, "build time environment variables to add/replace")
	cmd.Flags().StringArrayVarP(&factory.DeleteEnv, "delete-env", "", []string{}, "build time environment variables to remove")

	return cmd
}
