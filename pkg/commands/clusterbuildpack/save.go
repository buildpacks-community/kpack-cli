// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack

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
		Short: "Create or patch a cluster buildpack",
		Long: `Create or patch a cluster buildpack by providing command line arguments.
The cluster buildpack will be created only if it does not exist, otherwise it will be patched.

No defaults will be assumed for patches.
`,
		Example:      "kp clusterbuildpack save my-buildpack --image gcr.io/paketo-buildpacks/java",
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

			cbp, err := cs.KpackClient.KpackV1alpha2().ClusterBuildpacks().Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return create(ctx, name, flags, ch, cs, w)
			} else if err != nil {
				return err
			}

			return patch(ctx, cbp, flags, ch, cs, w)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "registry location where the buildpack is located")
	commands.SetDryRunOutputFlags(cmd)
	return cmd
}
