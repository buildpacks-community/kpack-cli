// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstore.Factory) *cobra.Command {
	var (
		buildpackages []string
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

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			factory.Printer = ch

			clusterStore, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(name, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return create(name, buildpackages, factory, ch, cs)
			} else if err != nil {
				return err
			}

			return update(clusterStore, buildpackages, factory, ch, cs)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	commands.SetDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	return cmd
}
