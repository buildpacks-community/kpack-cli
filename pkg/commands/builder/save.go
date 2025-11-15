// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags     CommandFlags
		tlsConfig registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or patch a builder",
		Long: `Create or patch a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace, otherwise it will be patched.

A buildpack order must be provided with either the path to an order yaml or via the --buildpack flag.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group. 

The --tag flag is required for a create but is immutable and will be ignored for a patch.

No defaults will be assumed for patches.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp builder save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
kp builder save my-builder --tag my-registry.com/my-builder-tag --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(flags.namespace)
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
			flags.namespace = cs.Namespace

			bldr, err := cs.KpackClient.KpackV1alpha2().Builders(cs.Namespace).Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				if flags.tag == "" {
					return errors.New("--tag is required to create the resource")
				}

				if flags.stack == "" {
					flags.stack = defaultStack
				}

				if flags.serviceAccount == "" {
					flags.serviceAccount = defaultServiceAccount
				}

				fetcher := registry.NewDefaultFetcher(tlsConfig)
			return create(ctx, name, flags, ch, cs, fetcher, w)
			} else if err != nil {
				return err
			}

			fetcher := registry.NewDefaultFetcher(tlsConfig)
		return patch(ctx, bldr, flags, ch, cs, fetcher, w)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", "", "stack resource to use (default \"default\" for a create)")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().StringSliceVarP(&flags.buildpacks, "buildpack", "b", []string{}, "buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'\n  repeat for each buildpack in order, or supply once with comma-separated list")
	cmd.Flags().StringVar(&flags.orderFrom, "order-from", "", "builder image to extract buildpack order from")
	cmd.Flags().StringVar(&flags.serviceAccount, "service-account", "", "service account name to use")
	commands.SetDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsConfig)
	return cmd
}
