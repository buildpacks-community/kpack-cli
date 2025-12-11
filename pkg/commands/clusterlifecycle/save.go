// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle

import (
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/clusterlifecycle"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/dockercreds"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		imageRef string
		tlsCfg   registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or patch a cluster lifecycle",
		Long: `Create or patch a cluster-scoped lifecycle by providing command line arguments.

The image will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.

Env vars can be used for registry auth as described in https://github.com/vmware-tanzu/kpack-cli/blob/main/docs/auth.md

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
The default service account used is read from the "default.repository.serviceaccount" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example:      `kp clusterlifecycle save my-lifecycle --image buildpacksio/lifecycle`,
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

			factory := clusterlifecycle.NewFactory(ch, rup.Relocator(ch.Writer(), tlsCfg, ch.IsUploading()), rup.Fetcher(tlsCfg))

			name := args[0]
			cLifecycle, err := cs.KpackClient.KpackV1alpha2().ClusterLifecycles().Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return create(ctx, name, imageRef, factory, ch, cs, w)
			} else if err != nil {
				return err
			}

			return patch(ctx, dockercreds.DefaultKeychain, cLifecycle, imageRef, factory, ch, cs, w)
		},
	}
	cmd.Flags().StringVarP(&imageRef, "image", "i", "", "image tag or local tar file path")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	_ = cmd.MarkFlagRequired("image")
	return cmd
}
