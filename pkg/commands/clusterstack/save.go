// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstack.Factory) *cobra.Command {
	var (
		buildImageRef string
		runImageRef   string
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Create or update a cluster stack",
		Long: `Create or update a cluster-scoped stack by providing command line arguments.

The run and build images will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.

The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
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

			name := args[0]
			factory.Printer = ch

			cStack, err := cs.KpackClient.KpackV1alpha1().ClusterStacks().Get(name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return create(name, buildImageRef, runImageRef, factory, ch, cs)
			} else if err != nil {
				return err
			}

			return update(cStack, buildImageRef, runImageRef, factory, ch, cs)
		},
	}
	cmd.Flags().StringVarP(&buildImageRef, "build-image", "b", "", "build image tag or local tar file path")
	cmd.Flags().StringVarP(&runImageRef, "run-image", "r", "", "run image tag or local tar file path")
	commands.SetDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")
	return cmd
}
