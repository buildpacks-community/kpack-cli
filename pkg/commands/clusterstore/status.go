// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"fmt"
	"io"
	"sort"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		verbose bool
	)

	cmd := &cobra.Command{
		Use:          "status <store-name>",
		Short:        "Display cluster store status",
		Long:         `Prints information about the status of a specific cluster-scoped store.`,
		Example:      "kp clusterstore status my-store",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			store, err := cs.KpackClient.KpackV1alpha2().ClusterStores().Get(cmd.Context(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			if verbose {
				return displayBuildpackagesDetailed(cmd.OutOrStdout(), store)
			} else {
				return displayBuildpackages(cmd.OutOrStdout(), store)
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "includes buildpacks and detection order")
	return cmd
}

type buildpackageInfo struct {
	id       string
	version  string
	homepage string
}

func displayStatus(out io.Writer, s *v1alpha2.ClusterStore) error {
	statusWriter := commands.NewStatusWriter(out)
	status := getStatusText(s)
	if err := statusWriter.AddBlock("", "Status", status); err != nil {
		return err
	}
	return statusWriter.Write()
}

func getStatusText(s *v1alpha2.ClusterStore) string {
	if cond := s.Status.GetCondition(corev1alpha1.ConditionReady); cond != nil {
		if cond.Status == corev1.ConditionTrue {
			return "Ready"
		} else if cond.Status == corev1.ConditionFalse {
			return "Not Ready - " + cond.Message
		}
	}
	return "Unknown"
}

func displayBuildpackages(out io.Writer, s *v1alpha2.ClusterStore) error {
	if err := displayStatus(out, s); err != nil {
		return err
	}

	buildpackages := getBuildpackageInfos(s)

	if len(buildpackages) <= 0 {
		return nil
	}

	writer, err := commands.NewTableWriter(out, "BUILDPACKAGE ID", "VERSION", "HOMEPAGE")
	if err != nil {
		return err
	}

	for _, buildpackage := range buildpackages {
		if err := writer.AddRow(buildpackage.id, buildpackage.version, buildpackage.homepage); err != nil {
			return err
		}
	}

	return writer.Write()
}

func displayBuildpackagesDetailed(out io.Writer, s *v1alpha2.ClusterStore) error {
	if err := displayStatus(out, s); err != nil {
		return err
	}

	buildpackages := map[string]corev1alpha1.BuildpackStatus{}
	buildpackageBps := map[string][]corev1alpha1.BuildpackStatus{}

	for _, b := range s.Status.Buildpacks {
		if b.Buildpackage.Id == "" && b.Buildpackage.Version == "" {
			continue
		}

		buildpackage := fmt.Sprintf("%s@%s", b.Buildpackage.Id, b.Buildpackage.Version)

		if b.Buildpackage.Id == b.Id && b.Buildpackage.Version == b.Version {
			buildpackages[buildpackage] = b
		} else {
			buildpackageBps[buildpackage] = append(buildpackageBps[buildpackage], b)
		}
	}

	return displayBuildpacks(out, buildpackages, buildpackageBps)
}

func getBuildpackageInfos(store *v1alpha2.ClusterStore) []buildpackageInfo {
	buildpackagesMap := make(map[string]buildpackageInfo)

	for _, buildpack := range store.Status.Buildpacks {
		if buildpack.Buildpackage.Id == "" && buildpack.Buildpackage.Version == "" {
			continue
		}

		buildpackageKey := fmt.Sprintf("%s@%s", buildpack.Buildpackage.Id, buildpack.Buildpackage.Version)
		if _, ok := buildpackagesMap[buildpackageKey]; !ok {
			buildpackagesMap[buildpackageKey] = buildpackageInfo{
				id:       buildpack.Buildpackage.Id,
				version:  buildpack.Buildpackage.Version,
				homepage: buildpack.Buildpackage.Homepage,
			}
		}
	}

	var buildpackageInfos []buildpackageInfo
	for _, buildpackageInfo := range buildpackagesMap {
		buildpackageInfos = append(buildpackageInfos, buildpackageInfo)
	}

	sort.Slice(buildpackageInfos, func(i, j int) bool {
		return buildpackageInfos[i].id < buildpackageInfos[j].id
	})

	return buildpackageInfos
}

func displayBuildpacks(out io.Writer, buildpackage map[string]corev1alpha1.BuildpackStatus, buildpacks map[string][]corev1alpha1.BuildpackStatus) error {
	var keys []string
	for k := range buildpackage {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		if i != 0 {
			_, _ = fmt.Fprintln(out, "")
		}

		statusWriter := commands.NewStatusWriter(out)

		err := statusWriter.AddBlock("",
			"Buildpackage", k,
			"Image", buildpackage[k].StoreImage.Image,
			"Homepage", buildpackage[k].Homepage,
		)
		if err != nil {
			return err
		}

		err = statusWriter.Write()
		if err != nil {
			return err
		}

		tbWriter, err := commands.NewTableWriter(out, "Buildpack id", "version", "homepage")
		if err != nil {
			return err
		}

		for _, bp := range buildpacks[k] {
			err = tbWriter.AddRow(bp.Id, bp.Version, bp.Homepage)
			if err != nil {
				return err
			}
		}

		err = tbWriter.Write()
		if err != nil {
			return err
		}

		orderTableWriter, err := commands.NewTableWriter(out, "Detection Order", "")
		if err != nil {
			return nil
		}

		for i, entry := range buildpackage[k].Order {
			err := orderTableWriter.AddRow(fmt.Sprintf("Group #%d", i+1), "")
			if err != nil {
				return err
			}
			for _, ref := range entry.Group {
				if ref.Optional {
					err := orderTableWriter.AddRow("  "+ref.Id, "(Optional)")
					if err != nil {
						return err
					}
				} else {
					err := orderTableWriter.AddRow("  "+ref.Id, "")
					if err != nil {
						return err
					}
				}
			}
		}

		err = orderTableWriter.Write()
		if err != nil {
			return err
		}
	}

	return nil
}
