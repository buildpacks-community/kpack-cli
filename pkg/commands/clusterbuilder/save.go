// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or patch a cluster builder",
		Long: `Create or patch a cluster builder by providing command line arguments.
The cluster builder will be created only if it does not exist, otherwise it is patched.

Tag when not specified, defaults to a combination of the canonical repository and specified builder name.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.

No defaults will be assumed for patches.
`,
		Example: `kp cb save my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb save my-builder --order /path/to/order.yaml
kp cb save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp cb save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]

			cb, err := cs.KpackClient.KpackV1alpha1().ClusterBuilders().Get(name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				if flags.stack == "" {
					flags.stack = defaultStack
				}

				if flags.store == "" {
					flags.store = defaultStore
				}

				return create(name, flags, ch, cs)
			} else if err != nil {
				return err
			}

			return patch(cb, flags, ch, cs)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use (default \"default\" for a create)")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use (default \"default\" for a create)")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&flags.output, "output", "", "output format. supported formats are: yaml, json")
	return cmd
}
