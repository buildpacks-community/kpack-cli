// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		buildpackages []string
		tlsCfg        registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "save <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Create or update a cluster store",
		Long: `Create or update a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

This clusterstore will be created only if it does not exist, otherwise it will be updated.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore save my-store -b my-registry.com/my-buildpackage
kp clusterstore save my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage
kp clusterstore save my-store -b ../path/to/my-local-buildpackage.cnb`,
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
			factory, err := clusterstore.NewFactory(ctx, cs, ch, rup, tlsCfg)
			if err != nil {
				return err
			}

			clusterStore, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return create(ctx, name, buildpackages, factory, ch, cs, w)
			} else if err != nil {
				return err
			}

			return update(ctx, clusterStore, buildpackages, factory, ch, cs, w)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}
