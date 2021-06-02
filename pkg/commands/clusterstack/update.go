// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"context"
	"io"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

type ImageFetcher interface {
	Fetch(src string) (v1.Image, error)
}

type ImageRelocator interface {
	Relocate(writer io.Writer, image v1.Image, dest string) (string, error)
}

func NewUpdateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		buildImageRef string
		runImageRef   string
		tlsCfg        registry.TLSConfig
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

			ctx := cmd.Context()

			stack, err := cs.KpackClient.KpackV1alpha1().ClusterStacks().Get(ctx, args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			factory := clusterstack.NewFactory(ch, rup.Relocator(ch.Writer(), tlsCfg, !ch.IsDryRun()), rup.Fetcher(tlsCfg))

			return update(ctx, authn.DefaultKeychain, stack, buildImageRef, runImageRef, factory, ch, cs, newWaiter(cs.DynamicClient))
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

func update(ctx context.Context, keychain authn.Keychain, stack *v1alpha1.ClusterStack, buildImageRef, runImageRef string, factory *clusterstack.Factory, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) error {
	if err := ch.PrintStatus("Updating ClusterStack..."); err != nil {
		return err
	}

	helper := k8s.DefaultConfigHelper(cs)
	kpConfig, err := helper.GetKpConfig(ctx)
	if err != nil {
		return err
	}

	hasUpdates, err := factory.UpdateStack(keychain, stack, buildImageRef, runImageRef, kpConfig)
	if err != nil {
		return err
	}

	if hasUpdates && !ch.IsDryRun() {
		stack, err = cs.KpackClient.KpackV1alpha1().ClusterStacks().Update(ctx, stack, metav1.UpdateOptions{})
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

	return ch.PrintChangeResult(hasUpdates, "ClusterStack %q updated", stack.Name)
}
