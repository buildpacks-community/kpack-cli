// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/image"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		namespace string
		subPath   string
		factory   image.Factory
		tlsCfg    registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "patch <name>",
		Short: "Patch an existing image resource",
		Long: `Patch an existing image resource by providing command line arguments.
This will fail if the image resource does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry as the existing image resource tag.
Therefore, you must have credentials to access the registry on your machine.

All tags found under Image.spec.additionalTags will be added to your built OCI image.
To append to the list of tags that will be added to a built image, use the "additional-tag" flag.
To remove a tag from the list of tags that will be added to a built image, use the "delete-additional-tag".
To replace the entire list of tags, use the "replace-additional-tag".

Environment variables may be provided by using the "--env" flag or deleted by using the "--delete-env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 --delete-env key3 --delete-env key3".

Service bindings may be provided by using the "--service-binding" flag or deleted by using the "--delete-service-binding" flag.
For each service binding, supply the "--service-binding" flag followed by the <KIND>:<APIVERSION>:<NAME> or just <NAME> which will default the kind to "Secret".
For example, "--service-binding my-secret-1 --service-binding CustomProvisionedService:v1beta1:my-ps" --delete-service-binding Secret:v1:my-secret-2

The --cache-size flag can only be used to increase the size of the existing cache.

Env vars can be used for registry auth as described in https://github.com/buildpacks-community/kpack-cli/blob/main/docs/auth.md
`,
		Example: `kp image patch my-image --git-revision my-other-branch
kp image patch my-image --blob https://my-blob-host.com/my-blob
kp image patch my-image --local-path /path/to/local/source/code
kp image patch my-image --local-path /path/to/local/source/code --builder my-builder
kp image patch my-image --env foo=bar --env color=red --delete-env apple --delete-env potato
kp image patch my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --service-binding my-secret --service-binding CustomProvisionedService:v1:my-ps --delete-service-binding Secret:v1:my-secret-2`,
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

			img, err := cs.KpackClient.KpackV1alpha2().Images(cs.Namespace).Get(ctx, args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			factory.SourceUploader = rup.SourceUploader(ch.Writer(), tlsCfg, ch.CanChangeState())
			factory.Printer = ch

			if cmd.Flag("sub-path").Changed {
				factory.SubPath = &subPath
			}

			wasPatched, img, err := patch(ctx, img, &factory, ch, cs)
			if err != nil {
				return err
			}

			if wasPatched && ch.ShouldWait() {
				_, err = newImageWaiter(cs).Wait(cmd.Context(), cmd.OutOrStdout(), img)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringArrayVar(&factory.AdditionalTags, "additional-tag", []string{}, "adds additional tags to push the OCI image to")
	cmd.Flags().StringArrayVar(&factory.ReplaceAdditionalTags, "replace-additional-tag", []string{}, "replaces all additional tags to push the OCI image to")
	cmd.Flags().StringArrayVar(&factory.DeleteAdditionalTags, "delete-additional-tag", []string{}, "additional tags to remove")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&factory.GitRepo, "git", "", "git repository url")
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision such as commit, tag, or branch (default \"main\")")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVar(&factory.Builder, "builder", "", "builder name")
	cmd.Flags().StringVar(&factory.ClusterBuilder, "cluster-builder", "", "cluster builder name")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "e", []string{}, "build time environment variables to add/replace")
	cmd.Flags().StringArrayVarP(&factory.DeleteEnv, "delete-env", "d", []string{}, "build time environment variables to remove")
	cmd.Flags().StringArrayVarP(&factory.ServiceBinding, "service-binding", "s", []string{}, "build time service bindings to add/replace")
	cmd.Flags().StringArrayVarP(&factory.DeleteServiceBinding, "delete-service-binding", "", []string{}, "build time service bindings to remove")
	cmd.Flags().StringVar(&factory.CacheSize, "cache-size", "", "cache size as a kubernetes quantity")
	cmd.Flags().StringVar(&factory.SuccessBuildHistoryLimit, "success-build-history-limit", "", "number of successful builds to keep, leave empty to use cluster default")
	cmd.Flags().StringVar(&factory.FailedBuildHistoryLimit, "failed-build-history-limit", "", "number of failed builds to keep, leave empty to use cluster default")
	cmd.Flags().StringVar(&factory.ServiceAccount, "service-account", "", "service account name to use")
	cmd.Flags().BoolP("wait", "w", false, "wait for image resource patch to be reconciled and tail resulting build logs")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}

func patch(ctx context.Context, img *v1alpha2.Image, factory *image.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (bool, *v1alpha2.Image, error) {
	if err := ch.PrintStatus("Patching Image Resource..."); err != nil {
		return false, nil, err
	}

	updatedImage, err := factory.UpdateImage(img)
	if err != nil {
		return false, nil, err
	}

	p, err := k8s.CreatePatch(img, updatedImage)
	if err != nil {
		return false, nil, err
	}

	hasPatch := len(p) > 0
	if hasPatch && !ch.IsDryRun() {
		updatedImage, err = cs.KpackClient.KpackV1alpha2().Images(cs.Namespace).Patch(ctx, img.Name, types.MergePatchType, p, metav1.PatchOptions{})
		if err != nil {
			return hasPatch, nil, err
		}
	}

	updatedImageArray := []runtime.Object{updatedImage}

	if err = ch.PrintObjs(updatedImageArray); err != nil {
		return hasPatch, nil, err
	}

	return hasPatch, updatedImage, ch.PrintChangeResult(hasPatch, fmt.Sprintf("Image Resource %q patched", img.Name))
}
