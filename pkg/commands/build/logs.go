package build

import (
	"context"
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/logs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/pivotal/build-service-cli/pkg/build"
)

func NewLogsCommand(kpackClient versioned.Interface, k8sClient k8s.Interface, defaultNamespace string) *cobra.Command {
	var (
		namespace   string
		buildNumber int
	)

	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "Tails build logs for an image",
		Long: `Tails the logs from the containers of a specified build of an image.
Defaults to tailing logs from the latest build if build is not specified`,
		Example:      "tbctl build logs my-image\ntbctl build logs my-image -b 2 -n my-namespace",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			buildList, err := kpackClient.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{
				LabelSelector: v1alpha1.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.Errorf("no builds for image \"%s\" found in \"%s\" namespace", args[0], namespace)
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))
				bld, err := findBuild(buildList, buildNumber, args[0], namespace)
				if err != nil {
					return err
				}
				return logs.NewBuildLogsClient(k8sClient).Tail(context.Background(), cmd.OutOrStdout(), args[0], bld.Labels[v1alpha1.BuildNumberLabel], namespace)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace")
	cmd.Flags().IntVarP(&buildNumber, "build", "b", -1, "build number")

	return cmd
}
