// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstack.Factory) *cobra.Command {
	var (
		buildImageRef string
		runImageRef   string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a cluster stack",
		Long: `Create a cluster-scoped stack by providing command line arguments.

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

			return create(name, buildImageRef, runImageRef, factory, ch, cs)
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

func create(name, buildImageRef, runImageRef string, factory *clusterstack.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (err error) {
	factory.Repository, err = k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	if err = ch.PrintStatus("Creating ClusterStack..."); err != nil {
		return err
	}

	stack, err := factory.MakeStack(name, buildImageRef, runImageRef)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		stack, err = cs.KpackClient.KpackV1alpha1().ClusterStacks().Create(stack)
		if err != nil {
			return err
		}
	}

	if err = ch.PrintObj(stack); err != nil {
		return err
	}

	return ch.PrintResult("ClusterStack %q created", stack.Name)
}
