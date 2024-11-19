// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/dockercreds"
	importpkg "github.com/buildpacks-community/kpack-cli/pkg/import"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

func NewImportCommand(
	differ importpkg.Differ,
	clientSetProvider k8s.ClientSetProvider,
	rup registry.UtilProvider,
	timestampProvider importpkg.TimestampProvider,
	confirmationProvider ConfirmationProvider,
	newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {

	var (
		filename    string
		showChanges bool
		force       bool
		tlsConfig   registry.TLSConfig
	)

	const (
		confirmMessage          = "Confirm with y:"
		noChangesConfirmMessage = "Re-upload images with y:"
	)

	confirmMsgMap := map[bool]string{
		true:  confirmMessage,
		false: noChangesConfirmMessage,
	}

	cmd := &cobra.Command{
		Use:   "import -f <filename>",
		Short: "Import dependencies for stores, stacks, and cluster builders",
		Long: `This operation will create or update clusterstores, clusterstacks, and clusterbuilders defined in the dependency descriptor.

kp import will always attempt to upload the stack, store, and builder images, even if the resources have not changed.
This can be used as a way to repair resources when registry images have been unexpectedly removed.

Env vars can be used for registry auth as described in https://github.com/buildpacks-community/kpack-cli/blob/main/docs/auth.md`,
		Example: `kp import -f dependencies.yaml
cat dependencies.yaml | kp import -f -`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			kpConfig := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx)

			imgFetcher := rup.Fetcher(tlsConfig)
			imgRelocator := rup.Relocator(ch.Writer(), tlsConfig, ch.CanChangeState())

			importer := importpkg.NewImporter(
				ch,
				cs.K8sClient,
				cs.KpackClient,
				imgFetcher,
				imgRelocator,
				newWaiter(cs.DynamicClient),
				timestampProvider,
			)

			rawDescriptor, err := readDescriptor(cmd, filename)
			if err != nil {
				return err
			}

			descriptor, err := importer.ReadDescriptor(rawDescriptor)
			if err != nil {
				return err
			}

			keychain := dockercreds.DefaultKeychain

			if showChanges {
				hasChanges, summary, err := importpkg.SummarizeChange(ctx, keychain, descriptor, kpConfig, importpkg.NewDefaultRelocatedImageProvider(imgFetcher), differ, cs)
				if err != nil {
					return err
				}

				err = ch.Printlnf(summary)
				if err != nil {
					return err
				}

				if !force {
					confirmed, err := confirmationProvider.Confirm(confirmMsgMap[hasChanges])
					if err != nil {
						return err
					}

					if !confirmed {
						return ch.Printlnf("Skipping import")
					}
				}
			}

			var objs []runtime.Object
			if ch.IsDryRun() {
				objs, err = importer.ImportDescriptorDryRun(
					ctx,
					keychain,
					kpConfig,
					rawDescriptor,
				)
				if err != nil {
					return err
				}
			} else {
				objs, err = importer.ImportDescriptor(
					ctx,
					keychain,
					kpConfig,
					rawDescriptor,
				)
				if err != nil {
					return err
				}
			}

			if err := ch.PrintObjs(objs); err != nil {
				return err
			}

			return ch.PrintResult("Imported resources")
		},
	}
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "dependency descriptor filename")
	cmd.Flags().BoolVar(&showChanges, "show-changes", false, "show a summary of resource changes before importing")
	cmd.Flags().BoolVar(&force, "force", false, "import without confirmation when showing changes")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsConfig)
	_ = cmd.MarkFlagRequired("filename")
	return cmd
}

func readDescriptor(cmd *cobra.Command, filename string) (string, error) {
	var (
		reader io.ReadCloser
		err    error
	)

	if filename == "-" {
		reader = io.NopCloser(cmd.InOrStdin())
	} else {
		reader, err = os.Open(filename)
		if err != nil {
			return "", err
		}
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
