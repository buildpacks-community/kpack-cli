// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle

import (
	"context"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/clusterlifecycle"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/dockercreds"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		imageRef string
		tlsCfg   registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a clusterlifecycle",
		Long: `Create a cluster-scoped lifecycle by providing command line arguments.

The image will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.

Env vars can be used for registry auth as described in https://github.com/vmware-tanzu/kpack-cli/blob/main/docs/auth.md

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
The default service account used is read from the "default.repository.serviceaccount" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterlifecycle create my-lifecycle --image buildpacksio/lifecycle
`,
		Args:         commands.ExactArgsWithUsage(1),
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

			ctx := cmd.Context()

			factory := clusterlifecycle.NewFactory(ch, rup.Relocator(ch.Writer(), tlsCfg, ch.IsUploading()), rup.Fetcher(tlsCfg))

			name := args[0]
			return create(ctx, name, imageRef, factory, ch, cs, newWaiter(cs.DynamicClient))
		},
	}
	cmd.Flags().StringVarP(&imageRef, "image", "i", "", "image tag or local tar file path")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	_ = cmd.MarkFlagRequired("image")
	return cmd
}

func create(ctx context.Context, name, imageRef string, factory *clusterlifecycle.Factory, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) (err error) {
	if err = ch.PrintStatus("Creating ClusterLifecycle..."); err != nil {
		return err
	}

	kpConfig := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx)

	lifecycle, err := factory.MakeLifecycle(dockercreds.DefaultKeychain, name, imageRef, kpConfig)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		lifecycle, err = cs.KpackClient.KpackV1alpha2().ClusterLifecycles().Create(ctx, lifecycle, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if err := w.Wait(ctx, lifecycle); err != nil {
			return err
		}
	}

	lifecycleArray := []runtime.Object{lifecycle}

	if err = ch.PrintObjs(lifecycleArray); err != nil {
		return err
	}

	return ch.PrintResult("ClusterLifecycle %q created", name)
}
