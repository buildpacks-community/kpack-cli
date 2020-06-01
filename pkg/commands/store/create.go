package store

import (
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/store"

	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/spf13/cobra"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *store.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <store> <buildpackage> [<buildpackage>...]",
		Short: "Create a store",
		Long: `Create a buildpack store by providing command line arguments.

Buildpackages will be uploaded to the the default registry configured on your store.
Therefore, you must have credentials to access the registry on your machine.

This store will be created only if it does not exist.`,
		Example: `tbctl store create my-store my-registry.com/my-buildpackage --default-repository some-registry.io/some-repo
tbctl store create my-store my-registry.com/my-buildpackage my-registry.com/my-other-buildpackage --default-repository some-registry.io/some-repo
tbctl store create my-store ../path/to/my-local-buildpackage.cnb --default-repository some-registry.io/some-repo`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			buildpackages := args[1:]
			factory.Printer = commands.NewPrinter(cmd)

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			newStore, err := factory.MakeStore(name, buildpackages...)
			if err != nil {
				return err
			}

			_, err = cs.KpackClient.ExperimentalV1alpha1().Stores().Create(newStore)
			if err != nil {
				return err
			}

			factory.Printer.Printf("\"%s\" created", newStore.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&factory.DefaultRepository, "default-repository", "", "the repository where the buildpackage images will be uploaded")
	_ = cmd.MarkFlagRequired("default-repository")

	return cmd
}
