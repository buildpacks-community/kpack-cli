// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

type ImageFetcher interface {
	Fetch(src string) (v1.Image, error)
}

type ImageRelocator interface {
	Relocate(writer io.Writer, image v1.Image, dest string) (string, error)
}

func NewUpdateCommand(clientSetProvider k8s.ClientSetProvider, factory *clusterstack.Factory) *cobra.Command {
	var (
		buildImageRef string
		runImageRef   string
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a cluster stack",
		Long: `Updates the run and build images of a specific cluster-scoped stack.

The run and build images will be uploaded to the the registry configured on your stack.
Therefore, you must have credentials to access the registry on your machine.`,
		Example: `kp clusterstack update my-stack --build-image my-registry.com/build --run-image my-registry.com/run
kp clusterstack update my-stack --build-image ../path/to/build.tar --run-image ../path/to/run.tar`,
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

			factory.Printer = ch

			stack, err := cs.KpackClient.KpackV1alpha1().ClusterStacks().Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			return update(stack, buildImageRef, runImageRef, factory, ch, cs)
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

func update(stack *v1alpha1.ClusterStack, buildImageRef, runImageRef string, factory *clusterstack.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (err error) {
	factory.Repository, err = k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	if err = ch.PrintStatus("Updating ClusterStack..."); err != nil {
		return err
	}

	hasUpdates, err := factory.UpdateStack(stack, buildImageRef, runImageRef)
	if err != nil {
		return err
	}

	if hasUpdates && !ch.IsDryRun() {
		stack, err = cs.KpackClient.KpackV1alpha1().ClusterStacks().Update(stack)
		if err != nil {
			return err
		}
	}

	if err = ch.PrintObj(stack); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasUpdates, "ClusterStack %q updated", stack.Name)
}
