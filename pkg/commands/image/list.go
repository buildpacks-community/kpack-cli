package image

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewListCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Display images in the desired namespace",
		Long:    "Prints a table of the most important information about images. You will only see images in your current namespace.\nIf no namespace is provided, the default namespace is queried.",
		Example: "tbctl image list\ntbctl image list -n my-namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			imageList, err := kpackClient.BuildV1alpha1().Images(namespace).List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(imageList.Items) == 0 {
				return errors.Errorf("no images found in \"%s\" namespace", namespace)
			} else {
				return displayImagesTable(cmd, imageList)
			}

		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

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
