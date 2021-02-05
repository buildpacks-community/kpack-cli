// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace     string
		allNamespaces bool
		filters       []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images",
		Long: `Prints a table of the most important information about images in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp image list
kp image list -A
kp image list -n my-namespace
kp image list --filter ready=true --filter latest-reason=commit,trigger`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			var imagesNamespace string

			if allNamespaces {
				imagesNamespace = ""
			} else {
				imagesNamespace = cs.Namespace
			}

			imageList, err := cs.KpackClient.KpackV1alpha1().Images(imagesNamespace).List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			imageList, err = filterImageList(imageList, filters)
			if err != nil {
				return err
			}

			if len(imageList.Items) == 0 {
				return errors.New("no images found")
			} else {
				return displayImagesTable(cmd, imageList)
			}

		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Return objects found in all namespaces")
	cmd.Flags().StringArrayVar(&filters, "filter", nil,
		`Each new filter argument requires an additional filter flag.
Multiple values can be provided using comma separation.
Supported filters and values:
  builder=string
  clusterbuilder=string
  latest-reason=commit,trigger,config,stack,buildpack
  ready=true,false,unknown`)

	return cmd
}

func displayImagesTable(cmd *cobra.Command, imageList *v1alpha1.ImageList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY", "LATEST REASON", "LATEST IMAGE", "NAMESPACE")
	if err != nil {
		return err
	}

	for _, img := range imageList.Items {
		err := writer.AddRow(img.Name, getReadyText(img), img.Status.LatestBuildReason, img.Status.LatestImage, img.Namespace)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getReadyText(img v1alpha1.Image) string {
	cond := img.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
