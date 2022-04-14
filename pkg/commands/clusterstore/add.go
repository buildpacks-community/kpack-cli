// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/kpack-cli/pkg/clusterstore"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewAddCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		buildpackages []string
		tlsCfg        registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "add <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Add buildpackage(s) to cluster store",
		Long: `Upload buildpackage(s) to a specific cluster-scoped buildpack store.

Buildpackages will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore add my-store -b my-registry.com/my-buildpackage
kp clusterstore add my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage -b my-registry.com/my-third-buildpackage
kp clusterstore add my-store -b ../path/to/my-local-buildpackage.cnb`,
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

			name := args[0]
			store, err := cs.KpackClient.KpackV1alpha2().ClusterStores().Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("ClusterStore '%s' does not exist", name)
			} else if err != nil {
				return err
			}

			relocator := rup.Relocator(ch.Writer(), tlsCfg, ch.IsUploading())
			fetcher := rup.Fetcher(tlsCfg)
			factory := clusterstore.NewFactory(ch, relocator, fetcher)

			return update(ctx, store, buildpackages, factory, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}

func update(ctx context.Context, store *v1alpha2.ClusterStore, buildpackages []string, factory *clusterstore.Factory, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) error {
	if err := ch.PrintStatus("Adding to ClusterStore..."); err != nil {
		return err
	}

	kpConfig := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx)

	updatedStore, err := factory.AddToStore(authn.DefaultKeychain, store, kpConfig, buildpackages...)
	if err != nil {
		return err
	}

	patch, err := k8s.CreatePatch(store, updatedStore)
	if err != nil {
		return err
	}

	hasUpdates := len(patch) > 0
	if hasUpdates && !ch.IsDryRun() {
		updatedStore, err = cs.KpackClient.KpackV1alpha2().ClusterStores().Update(ctx, updatedStore, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		if err := w.Wait(ctx, updatedStore); err != nil {
			return err
		}
	}

	if err = ch.PrintObj(updatedStore); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasUpdates, "ClusterStore %q updated", updatedStore.Name)
}
