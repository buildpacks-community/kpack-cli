package image

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images",
		Long: `Prints a table of the most important information about images from the provided namespace.

namespace defaults to the kubernetes current-context namespace.`,
		Example: "tbctl image list\ntbctl image list -n my-namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			imageList, err := cs.KpackClient.BuildV1alpha1().Images(cs.Namespace).List(metav1.ListOptions{})
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

	return cmd
}

func displayImagesTable(cmd *cobra.Command, imageList *v1alpha1.ImageList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY", "LATEST IMAGE")
	if err != nil {
		return err
	}

	for _, img := range imageList.Items {
		err := writer.AddRow(img.Name, getReadyText(img), img.Status.LatestImage)
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
