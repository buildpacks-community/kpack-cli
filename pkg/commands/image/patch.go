package image

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, factory *image.PatchFactory) *cobra.Command {
	var (
		namespace string
		subPath   string
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing image configuration",
		Long: `Patch an existing image configuration by providing command line arguments.
This will fail if the image does not already exist in the provided namespace.

namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry as the existing image tag.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".

Existing environment variables may be deleted by using the "--delete-env" flag.
For each environment variable, supply the "--delete-env" flag followed by the variable name.
For example, "--delete-env key1 --delete-env key2 ...".`,
		Example: `tbctl image patch my-image --git-revision my-other-branch
tbctl image patch my-image --blob https://my-blob-host.com/my-blob
tbctl image patch my-image --local-path /path/to/local/source/code
tbctl image patch my-image --local-path /path/to/local/source/code --builder my-builder
tbctl image patch my-image --env foo=bar --env color=red --delete-env apple --delete-env potato`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			img, err := cs.KpackClient.BuildV1alpha1().Images(cs.Namespace).Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			if cmd.Flag("sub-path").Changed {
				factory.SubPath = &subPath
			}

			patch, err := factory.MakePatch(img)
			if err != nil {
				return err
			}

			if len(patch) == 0 {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "nothing to patch")
				return err
			}

			_, err = cs.KpackClient.BuildV1alpha1().Images(cs.Namespace).Patch(args[0], types.MergePatchType, patch)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" patched\n", img.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&factory.GitRepo, "git", "", "git repository url")
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVar(&factory.Builder, "builder", "", "builder name")
	cmd.Flags().StringVar(&factory.ClusterBuilder, "cluster-builder", "", "cluster builder name")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "e", []string{}, "build time environment variables to add/replace")
	cmd.Flags().StringArrayVarP(&factory.DeleteEnv, "delete-env", "d", []string{}, "build time environment variables to remove")

	return cmd
}
