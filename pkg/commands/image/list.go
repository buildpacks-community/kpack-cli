package image

import (
	"fmt"
	"text/tabwriter"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "no images found in %s namespace\n", namespace)
				return err
			} else {
				return displayImagesTable(cmd, imageList)
			}

		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "the namespace to query")

	return cmd
}

func displayImagesTable(cmd *cobra.Command, imageList *v1alpha1.ImageList) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 4, ' ', 0)

	_, err := fmt.Fprintln(writer, "NAME\tREADY\tLATEST IMAGE")
	if err != nil {
		return err
	}

	for _, img := range imageList.Items {
		_, err = fmt.Fprintf(writer, "%s\t%s\t%s\n", img.Name, getReadyText(img), img.Status.LatestImage)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func getReadyText(img v1alpha1.Image) string {
	cond := img.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
