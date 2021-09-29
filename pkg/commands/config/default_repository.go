package config

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewDefaultRepositoryCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default-repository [url]",
		Short: "Set or Get the default repository",
		Long: `Set or Get the default repository 

The default repository is the location where imported and cluster-level resources are stored. It is required to be configured to use kp.

This data is stored in a config map in the kpack namespace called kp-config. 
The kp-config config map also contains a service account that contains the secrets required to write to the default repository.

If this config map doesn't exist, it will automatically be created by running this command, using the default service account in the kpack namespace as the default service account.
`,
		Example: `kp config default-repository
kp config default-repository my-registry.com/my-default-repo`,
		Args:         commands.OptionalArgsWithUsage(1),
		Aliases:      []string{"cr"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			configHelper := config.NewKpConfigProvider(cs.K8sClient)

			if len(args) == 0 {
				kpConfig := configHelper.GetKpConfig(ctx)

				repo, err := kpConfig.DefaultRepository()
				if err != nil {
					return err
				}

				return ch.Printlnf("%s", repo)
			}

			err = configHelper.SetDefaultRepository(ctx, args[0])
			if err != nil {
				return err
			}

			return ch.Printlnf("kp-config set")
		},
	}

	return cmd
}
