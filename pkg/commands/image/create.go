// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"encoding/json"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, factory *image.Factory, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		tag       string
		namespace string
		subPath   string
		wait      bool
		dryRun    bool
		output    string
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

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".`,
		Example: `kp image create my-image --tag my-registry.com/my-repo --git https://my-repo.com/my-app.git --git-revision my-branch
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob
kp image create my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code
kp image create my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code --builder my-builder -n my-namespace
kp image create my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --env food=apple`,
		Args:         cobra.ExactArgs(1),
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

			img, err := create(name, tag, factory, ch, cs)
			if err != nil {
				return err
			}

			if ch.CanWait() {
				_, err := newImageWaiter(cs).Wait(cmd.Context(), cmd.OutOrStdout(), img)
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
	cmd.Flags().StringVar(&factory.GitRevision, "git-revision", "", "git revision (default \"master\")")
	cmd.Flags().StringVar(&factory.Blob, "blob", "", "source code blob url")
	cmd.Flags().StringVar(&factory.LocalPath, "local-path", "", "path to local source code")
	cmd.Flags().StringVar(&subPath, "sub-path", "", "build code at the sub path located within the source code directory")
	cmd.Flags().StringVarP(&factory.Builder, "builder", "b", "", "builder name")
	cmd.Flags().StringVarP(&factory.ClusterBuilder, "cluster-builder", "c", "", "cluster builder name")
	cmd.Flags().StringArrayVar(&factory.Env, "env", []string{}, "build time environment variables")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for image create to be reconciled and tail resulting build logs")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&output, "output", "", "output format. supported formats are: yaml, json")

	cmd.MarkFlagRequired("tag")
	return cmd
}

func create(name, tag string, factory *image.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (*v1alpha1.Image, error) {
	img, err := factory.MakeImage(name, cs.Namespace, tag)
	if err != nil {
		return nil, err
	}

	imgConfig, err := json.Marshal(img)
	if err != nil {
		return nil, err
	}

	if img.Annotations == nil {
		img.Annotations = map[string]string{}
	}
	img.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = string(imgConfig)

    if !ch.IsDryRun() {
		img, err = cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Create(img)
		if err != nil {
			return nil, err
		}
	}

	err = ch.PrintObj(img)
	if err != nil {
		return nil, err
	}

	return img, ch.PrintResult("%q created", img.Name)
}
