// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/buildchange"
	"github.com/pivotal/kpack/pkg/differ"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/build"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

const (
	buildMetadataKey = "io.buildpacks.build.metadata"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider) *cobra.Command {
	var (
		namespace   string
		buildNumber string
	)

	cmd := &cobra.Command{
		Use:   "status <image-name>",
		Short: "Display status for an image resource build",
		Long: `Prints detailed information about the status of a specific build of an image resource in the provided namespace.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp build status my-image\nkp build status my-image -b 2 -n my-namespace",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			buildList, err := cs.KpackClient.KpackV1alpha2().Builds(cs.Namespace).List(cmd.Context(), metav1.ListOptions{
				LabelSelector: v1alpha2.ImageLabel + "=" + args[0],
			})
			if err != nil {
				return err
			}

			if len(buildList.Items) == 0 {
				return errors.New("no builds found")
			} else {
				sort.Slice(buildList.Items, build.Sort(buildList.Items))
				bld, err := findBuild(buildList, buildNumber)
				if err != nil {
					return err
				}

				return displayBuildStatus(cmd, bld)
			}
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&buildNumber, "build", "b", "", "build number")

	return cmd
}

func findBuild(buildList *v1alpha2.BuildList, buildNumberString string) (v1alpha2.Build, error) {

	if buildNumberString == "" {
		return buildList.Items[len(buildList.Items)-1], nil
	}

	buildNumber, err := strconv.Atoi(buildNumberString)
	if err != nil {
		return v1alpha2.Build{}, errors.Errorf("build number should be an integer: %v", buildNumberString)
	}

	for _, b := range buildList.Items {
		val, err := strconv.Atoi(b.Labels[v1alpha2.BuildNumberLabel])
		if err != nil {
			return v1alpha2.Build{}, err
		}

		if val == buildNumber {
			return b, nil
		}
	}

	return v1alpha2.Build{}, errors.Errorf("build \"%d\" not found", buildNumber)
}

func displayBuildStatus(cmd *cobra.Command, bld v1alpha2.Build) error {
	statusWriter := commands.NewStatusWriter(cmd.OutOrStdout())

	reason, err := buildReason(bld)
	if err != nil {
		return errors.Wrapf(err, "error printing build reason")
	}

	statusItems := []string{
		"Image", bld.Status.LatestImage,
		"Status", getStatus(bld),
		"Reason", reason,
	}

	cond := bld.Status.GetCondition(corev1alpha1.ConditionSucceeded)
	if cond != nil {
		if cond.Reason != "" {
			statusItems = append(statusItems, "Status Reason", cond.Reason)
		}
		if cond.Message != "" {
			statusItems = append(statusItems, "Status Message", cond.Message)
		}
	}

	err = statusWriter.AddBlock("", statusItems...)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"",
		"Started", getStarted(bld),
		"Finished", getFinished(bld),
	)

	err = statusWriter.AddBlock("",
		"Pod Name", bld.Status.PodName)
	if err != nil {
		return err
	}

	err = statusWriter.AddBlock(
		"",
		"Builder", bld.Spec.Builder.Image,
		"Run Image", bld.Status.Stack.RunImage,
	)
	if err != nil {
		return err
	}

	if bld.Spec.Source.Git != nil {
		err = statusWriter.AddBlock(
			"",
			"Source", "GitUrl",
			"Url", bld.Spec.Source.Git.URL,
			"Revision", bld.Spec.Source.Git.Revision,
		)
		if err != nil {
			return err
		}
	} else if bld.Spec.Source.Blob != nil {
		err = statusWriter.AddBlock(
			"",
			"Source", "Blob",
			"Url", bld.Spec.Source.Blob.URL,
		)
		if err != nil {
			return err
		}
	} else {
		err = statusWriter.AddBlock("", "Source", "Local Source")
		if err != nil {
			return err
		}
	}

	err = statusWriter.Write()
	if err != nil {
		return err
	}

	tableWriter, err := commands.NewTableWriter(cmd.OutOrStdout(), "Buildpack Id", "Buildpack Version", "Homepage")
	if err != nil {
		return err
	}

	for _, buildpack := range bld.Status.BuildMetadata {
		err := tableWriter.AddRow(buildpack.Id, buildpack.Version, buildpack.Homepage)
		if err != nil {
			return err
		}
	}

	return tableWriter.Write()
}

func buildReason(bld v1alpha2.Build) (string, error) {
	var err error
	var reasonsStr, changesStr string

	changes, ok := bld.Annotations[v1alpha2.BuildChangesAnnotation]
	if ok {
		reasonsStr, changesStr, err = reasonsAndChanges(changes)
		if err != nil {
			return "", errors.Wrapf(err, "error generating build reason from string '%s'", changes)
		}
	} else {
		reasonsStr = bld.Annotations[v1alpha2.BuildReasonAnnotation]
	}

	var buildReasonStr string
	if changesStr == "" {
		buildReasonStr = reasonsStr
	} else {
		buildReasonStr = fmt.Sprintf("%s\n%s", reasonsStr, changesStr)
	}

	return buildReasonStr, nil
}

func reasonsAndChanges(changesJson string) (string, string, error) {
	var changes []buildchange.GenericChange
	if err := json.Unmarshal([]byte(changesJson), &changes); err != nil {
		return "", "", err
	}

	var reasons []string
	var sb strings.Builder

	o := differ.DefaultOptions()
	o.Prefix = "\t"
	d := differ.NewDiffer(o)

	for _, change := range changes {
		reasons = append(reasons, change.Reason)
		diff, err := d.Diff(change.Old, change.New)
		if err != nil {
			return "", "", err
		}
		sb.WriteString(diff)
	}

	reasonsStr := strings.Join(reasons, ",")
	changesStr := strings.TrimSuffix(sb.String(), "\n")
	return reasonsStr, changesStr, nil
}
