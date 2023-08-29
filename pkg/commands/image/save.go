// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewSaveCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		tag       string
		namespace string
		subPath   string
		factory   image.Factory
		tlsCfg    registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "save <name> --tag <tag>",
		Short: "Create or patch an image resource",
		Long: `Create or patch an image resource by providing command line arguments.
This image resource will be created only if it does not exist in the provided namespace, otherwise it will be patched.

The --tag flag is required for a create but is immutable and will be ignored for a patch.
The --cache-size flag can only be used to create or increase the size of the existing cache.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry provided for the image resource tag.
--local-path-destination-image can be used to specify the repository of the source code image.
If not specified, the source code image will be pushed to the <image-tag-repo>-source repo.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag or deleted by using the "--delete-env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 --delete-env key3".

Service bindings may be provided by using the "--service-binding" flag or deleted by using the "--delete-service-binding" flag.
For each service binding, supply the "--service-binding" flag followed by the <KIND>:<APIVERSION>:<NAME> or just <NAME> which will default the kind to "Secret".
For example, "--service-binding my-secret-1 --service-binding CustomProvisionedService:v1beta1:my-ps --delete-service-binding Secret:v1:my-secret-2"

Env vars can be used for registry auth as described in https://github.com/vmware-tanzu/kpack-cli/blob/main/docs/auth.md
`,
		Example: `kp image create my-image --tag my-registry.com/my-repo --git https://my-repo.com/my-app.git --git-revision my-branch
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob
kp image save my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code
kp image save my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code --builder my-builder -n my-namespace
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --env food=apple --delete-env apple --delete-env potato
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --service-binding my-secret --service-binding CustomProvisionedService:v1:my-ps --delete-service-binding Secret:v1:my-secret-2`,
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
			shouldWait := ch.ShouldWait()

			factory.SourceUploader = rup.SourceUploader(ch.Writer(), tlsCfg, ch.CanChangeState())
			factory.Printer = ch

			ctx := cmd.Context()

			img, err := cs.KpackClient.KpackV1alpha2().Images(cs.Namespace).Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				if tag == "" {
					return errors.New("--tag is required to create the resource")
				}

				factory.SubPath = &subPath
				img, err = create(ctx, name, tag, &factory, ch, cs)
			} else if err != nil {
				return err
			} else {
				if cmd.Flag("sub-path").Changed {
					factory.SubPath = &subPath
				}

				var patched bool
				patched, img, err = patch(ctx, img, &factory, ch, cs)
				if !patched {
					shouldWait = false
				}
			}

			if err != nil {
				return err
			}

			if shouldWait {
				if _, err := newImageWaiter(cs).Wait(ctx, cmd.OutOrStdout(), img); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "registry location where the image will be created")
	cmd.Flags().StringArrayVar(&factory.AdditionalTags, "additional-tag", []string{}, "additional tags to push the OCI image to")
	cmd.Flags().StringArrayVar(&factory.DeleteAdditionalTags, "delete-additional-tag", []string{}, "additional tags to remove")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&factory.GitRepo, "git", "", "git repository url")
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision such as commit, tag, or branch (default \"main\")")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&factory.LocalPathDestinationImage, "local-path-destination-image", "", "registry location of where the local source code will be uploaded to (default \"<image-tag-repo>-source\")")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVar(&factory.CacheSize, "cache-size", "", "cache size as a kubernetes quantity (default \"2G\")")
	cmd.Flags().StringVar(&factory.SuccessBuildHistoryLimit, "success-build-history-limit", "", "set the successBuildHistoryLimit")
	cmd.Flags().StringVar(&factory.FailedBuildHistoryLimit, "failed-build-history-limit", "", "set the failedBuildHistoryLimit")
	cmd.Flags().StringVarP(&factory.Builder, "builder", "b", "", "builder name")
	cmd.Flags().StringVarP(&factory.ClusterBuilder, "cluster-builder", "c", "", "cluster builder name")
	cmd.Flags().StringArrayVarP(&factory.Env, "env", "e", []string{}, "build time environment variables")
	cmd.Flags().StringArrayVarP(&factory.DeleteEnv, "delete-env", "d", []string{}, "build time environment variables to remove")
	cmd.Flags().StringArrayVarP(&factory.ServiceBinding, "service-binding", "s", []string{}, "build time service bindings to add/replace")
	cmd.Flags().StringArrayVarP(&factory.DeleteServiceBinding, "delete-service-binding", "", []string{}, "build time service bindings to remove")
	cmd.Flags().StringVar(&factory.ServiceAccount, "service-account", "", "service account name to use")
	cmd.Flags().BoolP("wait", "w", false, "wait for image create to be reconciled and tail resulting build logs")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}
