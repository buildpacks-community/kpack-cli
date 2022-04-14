// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/clusterstack"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		buildImageRef string
		runImageRef   string
		tlsCfg        registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a cluster stack",
		Long: `Create a cluster-scoped stack by providing command line arguments.

The run and build images will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
The default service account used is read from the "default.serviceaccount" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstack create my-stack --build-image my-registry.com/build --run-image my-registry.com/run
kp clusterstack create my-stack --build-image ../path/to/build.tar --run-image ../path/to/run.tar`,
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

			factory := clusterstack.NewFactory(ch, rup.Relocator(ch.Writer(), tlsCfg, ch.IsUploading()), rup.Fetcher(tlsCfg))

			name := args[0]
			return create(ctx, name, buildImageRef, runImageRef, factory, ch, cs, newWaiter(cs.DynamicClient))
		},
	}
	cmd.Flags().StringVarP(&buildImageRef, "build-image", "b", "", "build image tag or local tar file path")
	cmd.Flags().StringVarP(&runImageRef, "run-image", "r", "", "run image tag or local tar file path")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")
	return cmd
}

func create(ctx context.Context, name, buildImageRef, runImageRef string, factory *clusterstack.Factory, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) (err error) {
	if err = ch.PrintStatus("Creating ClusterStack..."); err != nil {
		return err
	}

	kpConfig := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx)

	stack, err := factory.MakeStack(authn.DefaultKeychain, name, buildImageRef, runImageRef, kpConfig)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		stack, err = cs.KpackClient.KpackV1alpha2().ClusterStacks().Create(ctx, stack, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if err := w.Wait(ctx, stack); err != nil {
			return err
		}
	}

	if err = ch.PrintObj(stack); err != nil {
		return err
	}

	return ch.PrintResult("ClusterStack %q created", name)
}
