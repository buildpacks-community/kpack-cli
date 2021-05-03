// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"

	"github.com/pivotal/build-service-cli/pkg/buildpackage"
	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	importpkg "github.com/pivotal/build-service-cli/pkg/import"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/lifecycle"
	"github.com/pivotal/build-service-cli/pkg/registry"
	"github.com/pivotal/build-service-cli/pkg/stackimage"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

func NewImportCommand(
	differ importpkg.Differ,
	clientSetProvider k8s.ClientSetProvider,
	rup registry.UtilProvider,
	timestampProvider TimestampProvider,
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
This can be used as a way to repair resources when registry images have been unexpectedly removed.`,
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

			configHelper := k8s.DefaultConfigHelper(cs)

			descriptor, err := getDependencyDescriptor(cmd, filename)
			if err != nil {
				return err
			}

			repository, err := configHelper.GetCanonicalRepository(ctx)
			if err != nil {
				return err
			}

			serviceAccount, err := configHelper.GetCanonicalServiceAccount(ctx)
			if err != nil {
				return err
			}

			imgFetcher := rup.Fetcher()
			imgRelocator := rup.Relocator(ch.CanChangeState())

			storeFactory := &clusterstore.Factory{
				Uploader: &buildpackage.Uploader{
					Fetcher:   imgFetcher,
					Relocator: imgRelocator,
				},
				TLSConfig:  tlsConfig,
				Repository: repository,
				Printer:    ch,
			}

			stackFactory := &clusterstack.Factory{
				Uploader: &stackimage.Uploader{
					Fetcher:   imgFetcher,
					Relocator: imgRelocator,
				},
				TLSConfig:  tlsConfig,
				Repository: repository,
				Printer:    ch,
			}

			importer := importer{
				client:            cs.KpackClient,
				commandHelper:     ch,
				timestampProvider: timestampProvider,
				waiter:            newWaiter(cs.DynamicClient),
			}

			if showChanges {
				hasChanges, summary, err := importpkg.SummarizeChange(ctx, descriptor, storeFactory, stackFactory, differ, cs)
				if err != nil {
					return err
				}
				ch.Printlnf(summary)

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

			if descriptor.HasLifecycleImage() {
				cfg := lifecycle.ImageUpdaterConfig{
					DryRun:       ch.IsDryRun(),
					IOWriter:     ch.Writer(),
					ImgFetcher:   imgFetcher,
					ImgRelocator: imgRelocator,
					ClientSet:    cs,
					TLSConfig:    tlsConfig,
				}

				err = importer.importLifecycle(ctx, descriptor.GetLifecycleImage(), cfg)
				if err != nil {
					return err
				}
			}

			storeToGen, err := importer.importClusterStores(ctx, descriptor.ClusterStores, storeFactory)
			if err != nil {
				return err
			}

			stackToGen, err := importer.importClusterStacks(ctx, descriptor.GetClusterStacks(), stackFactory)
			if err != nil {
				return err
			}

			if err := importer.importClusterBuilders(ctx, descriptor.GetClusterBuilders(), repository, serviceAccount, storeToGen, stackToGen); err != nil {
				return err
			}

			if err := ch.PrintObjs(importer.objects()); err != nil {
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

func getDependencyDescriptor(cmd *cobra.Command, filename string) (importpkg.DependencyDescriptor, error) {
	var (
		reader io.ReadCloser
		err    error
	)

	if filename == "-" {
		reader = ioutil.NopCloser(cmd.InOrStdin())
	} else {
		reader, err = os.Open(filename)
		if err != nil {
			return importpkg.DependencyDescriptor{}, err
		}
	}
	defer reader.Close()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return importpkg.DependencyDescriptor{}, err
	}

	var api importpkg.API
	if err := yaml.Unmarshal(buf, &api); err != nil {
		return importpkg.DependencyDescriptor{}, err
	}

	var deps importpkg.DependencyDescriptor
	switch api.Version {
	case importpkg.APIVersionV1:
		var d1 importpkg.DependencyDescriptorV1
		if err := yaml.Unmarshal(buf, &d1); err != nil {
			return importpkg.DependencyDescriptor{}, err
		}
		deps = d1.ToNextVersion()
	case importpkg.CurrentAPIVersion:
		if err := yaml.Unmarshal(buf, &deps); err != nil {
			return importpkg.DependencyDescriptor{}, err
		}
	default:
		return importpkg.DependencyDescriptor{}, errors.Errorf("did not find expected apiVersion, must be one of: %s", []string{importpkg.APIVersionV1, importpkg.CurrentAPIVersion})
	}

	if err := deps.Validate(); err != nil {
		return importpkg.DependencyDescriptor{}, err
	}

	return deps, nil
}
