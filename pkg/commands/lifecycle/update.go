// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/lifecycle"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

func NewUpdateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider) *cobra.Command {
	var (
		image  string
		tlsCfg registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "update --image <image-tag>",
		Short: "Update lifecycle image used by kpack",
		Long: `Update lifecycle image used by kpack

The Lifecycle image will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

The canonical repository is read from the "canonical.repository" key of the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example:      "kp lifecycle update --image my-registry.com/lifecycle",
		Args:         commands.ExactArgsWithUsage(0),
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if image == "" {
				return fmt.Errorf("required flag(s) \"image\" not set\n\n%s", cmd.UsageString())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			if err = ch.PrintStatus("Updating lifecycle image..."); err != nil {
				return err
			}

			cfg := lifecycle.ImageUpdaterConfig{
				DryRun:       ch.IsDryRun(),
				IOWriter:     ch.Writer(),
				ImgFetcher:   rup.Fetcher(),
				ImgRelocator: rup.Relocator(ch.CanChangeState()),
				ClientSet:    cs,
				TLSConfig:    tlsCfg,
			}

			configMap, err := lifecycle.UpdateImage(image, cfg)
			if err != nil {
				return err
			}

			if err := ch.PrintObj(configMap); err != nil {
				return err
			}

			return ch.PrintResult("Updated lifecycle image")
		},
	}
	cmd.Flags().StringVarP(&image, "image", "i", "", "location of the image")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}
