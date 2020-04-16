package image

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDeleteCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete an image",
		Long:    "Delete an image and the associated image builds from the cluster.",
		Example: "tbctl image delete my-image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := kpackClient.BuildV1alpha1().Images(namespace).Delete(args[0], &metav1.DeleteOptions{})
			if err != nil {
				return errors.Errorf("image \"%s\" not found", args[0])
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" deleted\n", args[0])
			return err
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "the namespace of the image to delete")

	return cmd
}
