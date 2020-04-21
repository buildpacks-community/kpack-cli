package image

import (
	"fmt"
	"sort"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/build"
)

const BuildNeededAnnotation = "image.build.pivotal.io/additionalBuildNeeded"

func NewTriggerCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:     "trigger <name>",
		Short:   "Trigger an image build",
		Long:    "Trigger an image build using inputs identical to the last build.",
		Example: "tbctl image trigger my-image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			buildList, err := kpackClient.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))

				build := buildList.Items[len(buildList.Items)-1].DeepCopy()
				build.Annotations[BuildNeededAnnotation] = time.Now().String()
				_, err := kpackClient.BuildV1alpha1().Builds(namespace).Update(build)
				if err != nil {
					return err
				}

				_, err = fmt.Fprintf(cmd.OutOrStderr(), "\"%s\" triggered\n", args[0])
				return err
			}
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")

	return cmd
}
