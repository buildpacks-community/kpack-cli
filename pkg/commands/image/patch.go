// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewPatchCommand(clientSetProvider k8s.ClientSetProvider, factory *image.Factory, newImageWaiter func(k8s.ClientSet) ImageWaiter) *cobra.Command {
	var (
		namespace    string
		subPath      string
		wait         bool
		dryRunConfig DryRunConfig
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
For example, "--delete-env key1 --delete-env key2 ...".`,
		Example: `kp image patch my-image --git-revision my-other-branch
kp image patch my-image --blob https://my-blob-host.com/my-blob
kp image patch my-image --local-path /path/to/local/source/code
kp image patch my-image --local-path /path/to/local/source/code --builder my-builder
kp image patch my-image --env foo=bar --env color=red --delete-env apple --delete-env potato`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			factory.Printer = commands.NewPrinter(cmd)

			name := args[0]

			img, err := cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if cmd.Flag("sub-path").Changed {
				factory.SubPath = &subPath
			}

			dryRunConfig.writer = cmd.OutOrStdout()

			img, err = patch(img, factory, dryRunConfig, cs)
			if err != nil {
				return err
			}

			if wait && !dryRunConfig.dryRun {
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
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for image patch to be reconciled and tail resulting build logs")
	cmd.Flags().BoolVarP(&dryRunConfig.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVarP(&dryRunConfig.outputFormat, "output", "o", "yaml", "output format. supported formats are: yaml, json")

	return cmd
}

func patch(img *v1alpha1.Image, factory *image.Factory, drc DryRunConfig, cs k8s.ClientSet) (*v1alpha1.Image, error) {
	patch, err := factory.MakePatch(img)
	if err != nil {
		return nil, err
	}

	if len(patch) == 0 {
		factory.Printer.Printf("nothing to patch")
		return nil, err
	}

	if drc.dryRun {
		printer, err := commands.NewResourcePrinter(drc.outputFormat)
		if err != nil {
			return nil, err
		}
		return nil, printer.PrintObject(img, drc.writer)
	}

	img, err = cs.KpackClient.KpackV1alpha1().Images(cs.Namespace).Patch(img.Name, types.MergePatchType, patch)
	if err != nil {
		return nil, err
	}

	factory.Printer.Printf("\"%s\" patched", img.Name)
	return img, nil
}
