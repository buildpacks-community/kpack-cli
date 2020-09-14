// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		tag       string
		namespace string
		stack     string
		store     string
		order     string
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or patch a builder",
		Long: `Create or patch a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace, otherwise it will be patched.

The --tag flag is required for a create but is immutable and will be ignored for a patch.

No defaults will be assumed for patches.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp builder save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			name := args[0]

			bldr, err := cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Get(name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				if tag == "" {
					return errors.New("--tag is required to create the resource")
				}

				if stack == "" {
					stack = defaultStack
				}

				if store == "" {
					store = defaultStore
				}

				return create(name, tag, cs.Namespace, stack, store, order, cmd, cs)
			} else if err != nil {
				return err
			}

			return patch(bldr, tag, stack, store, order, cmd, cs)
		},
	}
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&stack, "stack", "s", "", "stack resource to use (default \"default\" for a create)")
	cmd.Flags().StringVar(&store, "store", "", "buildpack store to use (default \"default\" for a create)")
	cmd.Flags().StringVarP(&order, "order", "o", "", "path to buildpack order yaml")

	return cmd
}
