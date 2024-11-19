package config

import (
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

func NewDefaultServiceAccountCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "default-service-account [name]",
		Short: "Set or Get the default service account",
		Long: `Set or Get the default service account 

The default service account will be set as the service account on all cluster builders created with kp and the secrets on the service account will used to provide credentials to write cluster builder images.

This data is stored in a config map in the kpack namespace called kp-config. 
The kp-config config map also contains the default repository which is the location that imported and cluster-level resources are stored.

If this config map doesn't exist, it will automatically be created by running this command, but the default repository field will be empty.
`,
		Example: `kp config default-service-account
kp config default-service-account my-service-account
kp config default-service-account my-service-account --service-account-namespace default`,
		Args:         commands.OptionalArgsWithUsage(1),
		Aliases:      []string{"csa"},
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

				serviceAccount := kpConfig.ServiceAccount()

				return ch.Printlnf("Name: %s\nNamespace: %s", serviceAccount.Name, serviceAccount.Namespace)
			}

			err = configHelper.SetDefaultServiceAccount(ctx, corev1.ObjectReference{Name: args[0], Namespace: namespace})
			if err != nil {
				return err
			}

			return ch.Printlnf("kp-config set")
		},
	}

	cmd.Flags().StringVarP(&namespace, "service-account-namespace", "", "kpack", "namespace of default service account")

	return cmd
}
