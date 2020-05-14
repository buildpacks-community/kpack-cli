package store

import (
	"fmt"
	"io"
	"sort"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func NewStatusCommand(contextProvider commands.ContextProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Display store status",
		Long:         `Prints detailed information about the status of the store.`,
		Example:      "tbctl store status",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := contextProvider.GetContext()
			if err != nil {
				return err
			}

			store, err := context.KpackClient.ExperimentalV1alpha1().Stores().Get(DefaultStoreName, v1.GetOptions{})
			if err != nil {
				return err
			}
			return displayStoreStatus(cmd.OutOrStdout(), store)
		},
	}
	return cmd
}

type buildpackageDetails struct {
	buildpackage string
	image        string
	homepage     string
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
