// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/clusterlifecycle"
	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/dockercreds"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		imageRef string
		tlsCfg   registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:     "patch <name>",
		Aliases: []string{"update"},
		Short:   "Patch a clusterlifecycle",
		Long: `Patches the image of a specific cluster-scoped lifecycle.

The image will be uploaded to the registry configured on your lifecycle.
Therefore, you must have credentials to access the registry on your machine.

Env vars can be used for registry auth as described in https://github.com/buildpacks-community/kpack-cli/blob/main/docs/auth.md

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
The default service account used is read from the "default.repository.serviceaccount" key in the "kp-config" ConfigMap within "kpack" namespace.`,
		Example:      `kp clusterlifecycle patch my-lifecycle --image buildpacksio/lifecycle`,
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

			lifecycle, err := cs.KpackClient.KpackV1alpha2().ClusterLifecycles().Get(ctx, args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			factory := clusterlifecycle.NewFactory(ch, rup.Relocator(ch.Writer(), tlsCfg, ch.IsUploading()), rup.Fetcher(tlsCfg))

			return patch(ctx, dockercreds.DefaultKeychain, lifecycle, imageRef, factory, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&imageRef, "image", "i", "", "image tag or local tar file path")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	_ = cmd.MarkFlagRequired("image")
	return cmd
}

func patch(ctx context.Context, keychain authn.Keychain, lifecycle *v1alpha2.ClusterLifecycle, imageRef string, factory *clusterlifecycle.Factory, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) error {
	if err := ch.PrintStatus("Updating ClusterLifecycle..."); err != nil {
		return err
	}

	kpConfig := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx)

	updatedLifecycle, err := factory.UpdateLifecycle(keychain, lifecycle, imageRef, kpConfig)
	if err != nil {
		return err
	}

	p, err := k8s.CreatePatch(lifecycle, updatedLifecycle)
	if err != nil {
		return err
	}

	hasUpdates := len(p) > 0
	if hasUpdates && !ch.IsDryRun() {
		lifecycle, err = cs.KpackClient.KpackV1alpha2().ClusterLifecycles().Patch(ctx, updatedLifecycle.Name, types.MergePatchType, p, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		if err := w.Wait(ctx, updatedLifecycle); err != nil {
			return err
		}
	}

	updatedLifecycleArray := []runtime.Object{updatedLifecycle}

	if err = ch.PrintObjs(updatedLifecycleArray); err != nil {
		return err
	}

	return ch.PrintChangeResult(hasUpdates, "ClusterLifecycle %q updated", updatedLifecycle.Name)
}
