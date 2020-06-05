package stack

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/stack"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *stack.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a stack",
		Long: `Create a stack by providing command line arguments.

The run and build images will be uploaded to the the registry provided by "--default-repository".
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.`,
		Example: `kp stack create my-stack --default-repository some-registry.io/some-repo --build-image my-registry.com/build --run-image my-registry.com/run
kp stack create my-stack --default-repository some-registry.io/some-repo --build-image ../path/to/build.tar --run-image ../path/to/run.tar`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			stk, err := factory.MakeStack(args[0])
			if err != nil {
				return err
			}

			_, err = cs.KpackClient.ExperimentalV1alpha1().Stacks().Create(stk)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", stk.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&factory.DefaultRepository, "default-repository", "", "", "the repository where the stack images will be relocated")
	cmd.Flags().StringVarP(&factory.BuildImageRef, "build-image", "", "", "build image tag or local tar file path")
	cmd.Flags().StringVarP(&factory.RunImageRef, "run-image", "", "", "run image tag or local tar file path")
	_ = cmd.MarkFlagRequired("default-repository")
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")

	return cmd
}
