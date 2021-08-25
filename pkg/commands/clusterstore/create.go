// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/clusterstore"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		buildpackages []string
		tlsCfg        registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "create <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Create a cluster store",
		Long: `Create a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

This clusterstore will be created only if it does not exist.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore create my-store -b my-registry.com/my-buildpackage
kp clusterstore create my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage
kp clusterstore create my-store -b ../path/to/my-local-buildpackage.cnb`,
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

			factory := clusterstore.NewFactory(ch, rup.Relocator(ch.Writer(), tlsCfg, ch.IsUploading()), rup.Fetcher(tlsCfg))

			name := args[0]
			return create(ctx, name, buildpackages, factory, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}

func create(ctx context.Context, name string, buildpackages []string, factory *clusterstore.Factory, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) (err error) {
	if err = ch.PrintStatus("Creating ClusterStore..."); err != nil {
		return err
	}

	kpConfig := config.NewKpConfigProvider(cs).GetKpConfig(ctx)

	newStore, err := factory.MakeStore(authn.DefaultKeychain, name, kpConfig, buildpackages...)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		newStore, err = cs.KpackClient.KpackV1alpha2().ClusterStores().Create(ctx, newStore, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if err := w.Wait(ctx, newStore); err != nil {
			return err
		}
	}

	if err = ch.PrintObj(newStore); err != nil {
		return err
	}

	return ch.PrintResult("ClusterStore %q created", name)
}
