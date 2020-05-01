package stack

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/stack"
)

func NewCreateCommand(kpackClient versioned.Interface, factory *stack.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a stack configuration",
		Long: `Create a stack configuration by providing command line arguments.
This stack will be created if it does not yet exist.

The flags for this command determine how the build will retrieve source code:

	"--git" and "--git-revision" to use Git based source

	"--blob" to use source code hosted in a blob store

	"--local-path" to use source code from the local machine

Local source code will be pushed to the same registry provided for the image tag.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by
the key value pair. For example, "--env key1=value1 --env key=value2 ...".
`,
		Example: `tbctl stack create my-stack --repo my-registry.com/my-repo --run-image gcr.io/paketo-buildpacks/run:base --build-image gcr.io/paketo-buildpacks/build:base
tbctl stack create my-stack --repo my-registry.com/my-repo --run-image /Users/home/userthis/images/run.tar --build-image /Users/home/userthis/images/build.tar
tbctl stack create my-stack --repo my-registry.com/my-repo --run-image gcr.io/paketo-buildpacks/run:base --build-image /Users/home/userthis/images/build.tar`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			stack, err := factory.MakeStack(args[0])
			if err != nil {
				return err
			}

			_, err = kpackClient.ExperimentalV1alpha1().Stacks().Create(stack)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", stack.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&factory.DefaultRepo, "repo", "", "", "repository to relocate stack images to. NOTE: your cluster needs the credentials required to access images in this docker repo.")
	cmd.Flags().StringVarP(&factory.BuildImage, "build-image", "b", "", "build image tag or local tar file path")
	cmd.Flags().StringVarP(&factory.RunImage, "run-image", "r", "", "build image tag or local tar file path")
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}
