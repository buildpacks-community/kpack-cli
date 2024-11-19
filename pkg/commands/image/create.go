// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/image"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		tag       string
		namespace string
		subPath   string
		factory   image.Factory
		tlsCfg    registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "create <name> --tag <tag>",
		Short: "Create an image resource",
		Long: `Create an image resource by providing command line arguments.
This image resource will be created only if it does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry provided for the image resource tag.
--local-path-destination-image can be used to specify the repository of the source code image.
If not specified, the source code image will be pushed to the <image-tag-repo>-source repo.
Therefore, you must have credentials to access the registry on your machine.
--registry-ca-cert-path and --registry-verify-certs are only used for local source type.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".

Service bindings may be provided by using the "--service-binding" flag.
For each service binding, supply the "--service-binding" flag followed by the <KIND>:<APIVERSION>:<NAME> or just <NAME> which will default the kind to "Secret".
For example, "--service-binding my-secret-1 --service-binding Secret:v1:my-secret-2 --service-binding CustomProvisionedService:v1beta1:my-ps

Env vars can be used for registry auth as described in https://github.com/buildpacks-community/kpack-cli/blob/main/docs/auth.md"`,
		Example: `kp image create my-image --tag my-registry.com/my-repo --git https://my-repo.com/my-app.git --git-revision my-branch
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob
kp image create my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code
kp image create my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code --builder my-builder -n my-namespace
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --env food=apple
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --service-binding my-secret-1 --service-binding Secret:v1:my-secret-2 --service-binding CustomProvisionedService:v1beta1:my-ps`,
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

			name := args[0]

			factory.SubPath = &subPath
			factory.SourceUploader = rup.SourceUploader(ch.Writer(), tlsCfg, ch.IsUploading())
			factory.Printer = ch

			ctx := cmd.Context()
			img, err := create(ctx, name, tag, &factory, ch, cs)
			if err != nil {
				return err
			}

			if ch.ShouldWait() {
				_, err := newImageWaiter(cs).Wait(ctx, cmd.OutOrStdout(), img)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "registry location where the OCI image will be created")
	cmd.Flags().StringArrayVar(&factory.AdditionalTags, "additional-tag", []string{}, "additional tags to push the OCI image to")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&factory.GitRepo, "git", "", "git repository url")
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision such as commit, tag, or branch (default \"main\")")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&factory.LocalPathDestinationImage, "local-path-destination-image", "", "registry location of where the local source code will be uploaded to (default \"<image-tag-repo>-source\")")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVarP(&factory.Builder, "builder", "b", "", "builder name")
	cmd.Flags().StringVarP(&factory.ClusterBuilder, "cluster-builder", "c", "", "cluster builder name")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "e", []string{}, "build time environment variables")
	cmd.Flags().StringArrayVarP(&factory.ServiceBinding, "service-binding", "s", []string{}, "build time service bindings")
	cmd.Flags().StringVar(&factory.CacheSize, "cache-size", "", "cache size as a kubernetes quantity (default \"2G\")")
	cmd.Flags().StringVar(&factory.SuccessBuildHistoryLimit, "success-build-history-limit", "", "number of successful builds to keep, leave empty to use cluster default")
	cmd.Flags().StringVar(&factory.FailedBuildHistoryLimit, "failed-build-history-limit", "", "number of failed builds to keep, leave empty to use cluster default")
	cmd.Flags().StringVar(&factory.ServiceAccount, "service-account", "default", "service account name to use")
	cmd.Flags().BoolP("wait", "w", false, "wait for image create to be reconciled and tail resulting build logs")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func create(ctx context.Context, name, tag string, factory *image.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (*v1alpha2.Image, error) {
	if err := ch.PrintStatus("Creating Image Resource..."); err != nil {
		return nil, err
	}

	img, err := factory.MakeImage(name, cs.Namespace, tag)
	if err != nil {
		return nil, err
	}

	if err := k8s.SetLastAppliedCfg(img); err != nil {
		return nil, err
	}

	if !ch.IsDryRun() {
		img, err = cs.KpackClient.KpackV1alpha2().Images(cs.Namespace).Create(ctx, img, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	imgArray := []runtime.Object{img}

	err = ch.PrintObjs(imgArray)
	if err != nil {
		return nil, err
	}

	return img, ch.PrintResult("Image Resource %q created", img.Name)
}
