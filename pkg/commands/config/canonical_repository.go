package config

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewCanonicalRepositoryCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "canonical-repository [url]",
		Short: "Set or Get the canonical repository",
		Long: `Set or Get the canonical repository 

The canonical repository is the location where imported and cluster-level resources are stored. It is required to be configured to use kp.

This data is stored in a config map in the kpack namespace called kp-config. 
The kp-config config map also contains a service account that contains the secrets required to write to the canonical repository.

If this config map doesn't exist, it will automatically be created by running this command, using the default service account in the kpack namespace as the canonical service account.
`,
		Example: `kp config canonical-repository
kp config canonical-repository my-registry.com/my-canonical-repo`,
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

			configHelper := config.NewKpConfigProvider(cs)

			if len(args) == 0 {
				kpConfig := configHelper.GetKpConfig(ctx)

				repo, err := kpConfig.CanonicalRepository()
				if err != nil {
					return err
				}

				return ch.Printlnf("%s", repo)
			}

			err = configHelper.SetCanonicalRepository(ctx, args[0])
			if err != nil {
				return err
			}

			return ch.Printlnf("kp-config set")
		},
	}

	return cmd
}
