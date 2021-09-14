// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or patch a cluster builder",
		Long: `Create or patch a cluster builder by providing command line arguments.
The cluster builder will be created only if it does not exist, otherwise it is patched.

A buildpack order must be provided with either the path to an order yaml or via the --buildpack flag.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group. 

Tag when not specified, defaults to a combination of the default repository and specified builder name.
The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.

No defaults will be assumed for patches.
`,
		Example: `kp cb save my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb save my-builder --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1
kp cb save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp cb save my-builder --tag my-registry.com/my-builder-tag --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			w := newWaiter(cs.DynamicClient)

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			cb, err := cs.KpackClient.KpackV1alpha2().ClusterBuilders().Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				if flags.stack == "" {
					flags.stack = defaultStack
				}

				if flags.store == "" {
					flags.store = defaultStore
				}
				return create(ctx, name, flags, ch, cs, w)
			} else if err != nil {
				return err
			}

			return patch(ctx, cb, flags, ch, cs, w)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use (default \"default\" for a create)")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use (default \"default\" for a create)")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().StringSliceVarP(&flags.buildpacks, "buildpack", "b", []string{}, "buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'\n  repeat for each buildpack in order, or supply once with comma-separated list")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}
