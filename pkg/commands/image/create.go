// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		tag       string
		namespace string
		subPath   string
		factory   image.Factory
	)

	cmd := &cobra.Command{
		Use:   "create <name> --tag <tag>",
		Short: "Create an image configuration",
		Long: `Create an image configuration by providing command line arguments.
This image will be created only if it does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry provided for the image tag.
Therefore, you must have credentials to access the registry on your machine.
--registry-ca-cert-path and --registry-verify-certs are only used for local source type.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".`,
		Example: `kp image create my-image --tag my-registry.com/my-repo --git https://my-repo.com/my-app.git --git-revision my-branch
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob
kp image create my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code
kp image create my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code --builder my-builder -n my-namespace
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --env food=apple`,
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
			factory.SourceUploader = rup.SourceUploader(ch.CanChangeState())
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
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "registry location where the image will be created")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVar(&factory.GitRepo, "git", "", "git repository url")
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision such as commit, tag, or branch (default \"main\")")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVarP(&factory.Builder, "builder", "b", "", "builder name")
	cmd.Flags().StringVarP(&factory.ClusterBuilder, "cluster-builder", "c", "", "cluster builder name")
	cmd.Flags().StringArrayVar(&factory.Env, "env", []string{}, "build time environment variables")
	cmd.Flags().StringVar(&factory.CacheSize, "cache-size", "", "cache size as a kubernetes quantity (default \"2G\")")
	cmd.Flags().BoolP("wait", "w", false, "wait for image create to be reconciled and tail resulting build logs")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &factory.TLSConfig)
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func create(ctx context.Context, name, tag string, factory *image.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (*v1alpha1.Image, error) {
	if err := ch.PrintStatus("Creating Image..."); err != nil {
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
		img, err = cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Create(ctx, img, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	err = ch.PrintObj(img)
	if err != nil {
		return nil, err
	}

	return img, ch.PrintResult("Image %q created", img.Name)
}
