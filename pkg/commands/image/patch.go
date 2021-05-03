// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		namespace string
		subPath   string
		factory   image.Factory
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing image configuration",
		Long: `Patch an existing image configuration by providing command line arguments.
This will fail if the image does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry as the existing image tag.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".

Existing environment variables may be deleted by using the "--delete-env" flag.
For each environment variable, supply the "--delete-env" flag followed by the variable name.
For example, "--delete-env key1 --delete-env key2 ...".

The --cache-size flag can only be used to increase the size of the existing cache.
`,
		Example: `kp image patch my-image --git-revision my-other-branch
kp image patch my-image --blob https://my-blob-host.com/my-blob
kp image patch my-image --local-path /path/to/local/source/code
kp image patch my-image --local-path /path/to/local/source/code --builder my-builder
kp image patch my-image --env foo=bar --env color=red --delete-env apple --delete-env potato`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			img, err := cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Get(ctx, args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			factory.SourceUploader = rup.SourceUploader(ch.CanChangeState())
			factory.Printer = ch

			if cmd.Flag("sub-path").Changed {
				factory.SubPath = &subPath
			}

			patched, img, err := patch(ctx, img, &factory, ch, cs)
			if err != nil {
				return err
			}

			if patched && ch.ShouldWait() {
				_, err = newImageWaiter(cs).Wait(cmd.Context(), cmd.OutOrStdout(), img)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&factory.GitRepo, "git", "", "git repository url")
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision (default \"master\")")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVar(&factory.Builder, "builder", "", "builder name")
	cmd.Flags().StringVar(&factory.ClusterBuilder, "cluster-builder", "", "cluster builder name")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "e", []string{}, "build time environment variables to add/replace")
	cmd.Flags().StringArrayVarP(&factory.DeleteEnv, "delete-env", "d", []string{}, "build time environment variables to remove")
	cmd.Flags().StringVar(&factory.CacheSize, "cache-size", "", "cache size as a kubernetes quantity")
	cmd.Flags().BoolP("wait", "w", false, "wait for image patch to be reconciled and tail resulting build logs")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	return cmd
}

func patch(ctx context.Context, img *v1alpha1.Image, factory *image.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (bool, *v1alpha1.Image, error) {
	if err := ch.PrintStatus("Patching Image..."); err != nil {
		return false, nil, err
	}

	patchedImage, patch, err := factory.MakePatch(img)
	if err != nil {
		return false, nil, err
	}

	hasPatch := len(patch) > 0
	if hasPatch && !ch.IsDryRun() {
		patchedImage, err = cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Patch(ctx, img.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			return hasPatch, nil, err
		}
	}

	if err = ch.PrintObj(patchedImage); err != nil {
		return hasPatch, nil, err
	}

	return hasPatch, patchedImage, ch.PrintChangeResult(hasPatch, fmt.Sprintf("Image %q patched", img.Name))
}
