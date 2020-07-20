// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"fmt"
	"io"
	"sort"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewStatusCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		verbose bool
	)

	cmd := &cobra.Command{
		Use:          "status <store-name>",
		Short:        "Display store status",
		Long:         `Prints information about the status of a specific store.`,
		Example:      "kp store status my-store",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			store, err := cs.KpackClient.ExperimentalV1alpha1().Stores().Get(args[0], v1.GetOptions{})
			if err != nil {
				return err
			}

			if verbose {
				return displayStoreStatus(cmd.OutOrStdout(), store)
			} else {
				return displayBuildpackagesTable(cmd.OutOrStdout(), getBuildpackageInfos(store))
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "includes buildpacks and detection order")
	return cmd
}

type buildpackageInfo struct {
	id      string
	version string
}

func getBuildpackageInfos(store *expv1alpha1.Store) []buildpackageInfo {
	buildpackagesMap := make(map[string]buildpackageInfo)

	for _, buildpack := range store.Status.Buildpacks {
		if buildpack.Buildpackage.Id == "" && buildpack.Buildpackage.Version == "" {
			continue
		}

		buildpackageKey := fmt.Sprintf("%s@%s", buildpack.Buildpackage.Id, buildpack.Buildpackage.Version)
		if _, ok := buildpackagesMap[buildpackageKey]; !ok {
			buildpackagesMap[buildpackageKey] = buildpackageInfo{
				id:      buildpack.Buildpackage.Id,
				version: buildpack.Buildpackage.Version,
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

func displayBuildpackagesTable(out io.Writer, buildpackages []buildpackageInfo) error {
	if len(buildpackages) <= 0 {
		return nil
	}

	writer, err := commands.NewTableWriter(out, "BUILDPACKAGE ID", "VERSION")
	if err != nil {
		return err
	}

	for _, buildpackage := range buildpackages {
		if err := writer.AddRow(buildpackage.id, buildpackage.version); err != nil {
			return err
		}
	}

	return writer.Write()
}

func displayStoreStatus(out io.Writer, s *expv1alpha1.Store) error {
	buildpackages := map[string]expv1alpha1.StoreBuildpack{}
	buildpackageBps := map[string][]expv1alpha1.StoreBuildpack{}

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

func displayBuildpacks(out io.Writer, buildpackage map[string]expv1alpha1.StoreBuildpack, buildpacks map[string][]expv1alpha1.StoreBuildpack) error {
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

		tbWriter, err := commands.NewTableWriter(out, "Buildpack id", "version")
		if err != nil {
			return err
		}

		for _, bp := range buildpacks[k] {
			err = tbWriter.AddRow(bp.Id, bp.Version)
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
