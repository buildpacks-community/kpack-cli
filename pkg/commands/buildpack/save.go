// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack

import (
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or patch a buildpack",
		Long: `Create or patch a buildpack by providing command line arguments.
The buildpack will be created only if it does not exist in the provided namespace, otherwise it will be patched.

No defaults will be assumed for patches.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp buildpack save my-buildpack --image gcr.io/paketo-buildpacks/java
kp buildpack save my-buildpack --service-account my-other-sa`,
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

			bp, err := cs.KpackClient.KpackV1alpha2().Buildpacks(cs.Namespace).Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				if flags.serviceAccount == "" {
					flags.serviceAccount = defaultServiceAccount
				}

				return create(ctx, name, flags, ch, cs, w)
			} else if err != nil {
				return err
			}

			return patch(ctx, bp, flags, ch, cs, w)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "registry location where the buildpack is located")
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&flags.serviceAccount, "service-account", "", "service account name to use")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}
