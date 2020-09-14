// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstack.Factory) *cobra.Command {
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
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			name := args[0]

			factory.Printer = commands.NewPrinter(cmd)

			return create(name, factory, cs)
		},
	}
	cmd.Flags().StringVarP(&factory.BuildImageRef, "build-image", "b", "", "build image tag or local tar file path")
	cmd.Flags().StringVarP(&factory.RunImageRef, "run-image", "r", "", "run image tag or local tar file path")
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")

	return cmd
}

func create(name string, factory *clusterstack.Factory, cs k8s.ClientSet) (err error) {
	factory.Repository, err = k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	stack, err := factory.MakeStack(name)
	if err != nil {
		return err
	}

	_, err = cs.KpackClient.KpackV1alpha1().ClusterStacks().Create(stack)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.Printer.Writer, "\"%s\" created\n", stack.Name)
	return err
}
