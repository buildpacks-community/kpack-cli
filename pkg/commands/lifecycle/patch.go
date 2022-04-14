// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/lifecycle"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func NewUpdateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider) *cobra.Command {
	var (
		image  string
		tlsCfg registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:     "patch --image <image-tag>",
		Aliases: []string{"update"},
		Short:   "Patch lifecycle image used by kpack",
		Long: `Patch lifecycle image used by kpack

The Lifecycle image will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.

The default repository is read from the "default.repository" key of the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example:      "kp lifecycle patch --image my-registry.com/lifecycle",
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

			if err = ch.PrintStatus("Patching lifecycle config..."); err != nil {
				return err
			}

			cfg := lifecycle.ImageUpdaterConfig{
				DryRun:       ch.IsDryRun(),
				IOWriter:     ch.Writer(),
				ImgFetcher:   rup.Fetcher(tlsCfg),
				ImgRelocator: rup.Relocator(ch.Writer(), tlsCfg, ch.CanChangeState()),
				ClientSet:    cs,
				TLSConfig:    tlsCfg,
			}

			configMap, err := lifecycle.UpdateImage(cmd.Context(), authn.DefaultKeychain, image, cfg)
			if err != nil {
				return err
			}

			if err := ch.PrintObj(configMap); err != nil {
				return err
			}

			return ch.PrintResult("Patched lifecycle config")
		},
	}
	cmd.Flags().StringVarP(&image, "image", "i", "", "location of the image")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}
