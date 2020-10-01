// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"github.com/ghodss/yaml"
	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	importpkg "github.com/pivotal/build-service-cli/pkg/import"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

func NewImportCommand(
	clientSetProvider k8s.ClientSetProvider,
	importFactory *importpkg.Factory,
	storeFactory *clusterstore.Factory,
	stackFactory *clusterstack.Factory,
	differ commands.Differ,
	confirmationProvider ConfirmationProvider) *cobra.Command {

	var (
		filename string
		dryRun   bool
		output   string
		force    bool
	)

	const (
		confirmMessage = "Confirm with y:"
	)

	cmd := &cobra.Command{
		Use:   "import -f <filename>",
		Short: "Import dependencies for stores, stacks, and cluster builders",
		Long:  `This operation will create or update stores, stacks, and cluster builders defined in the dependency descriptor.`,
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

			configHelper := k8s.DefaultConfigHelper(cs)

			descriptor, err := getDependencyDescriptor(cmd, filename)
			if err != nil {
				return err
			}

			repository, err := configHelper.GetCanonicalRepository()
			if err != nil {
				return err
			}

			serviceAccount, err := configHelper.GetCanonicalServiceAccount()
			if err != nil {
				return err
			}

			storeFactory.Repository = repository // FIXME
			storeFactory.Printer = ch

			stackFactory.Repository = repository
			stackFactory.Printer = ch
			stackFactory.TLSConfig = storeFactory.TLSConfig

			importFactory.Client = cs.KpackClient

			var sBuilder strings.Builder
			sBuilder.WriteString("Cluster Stores\n\n")
			for _, c := range descriptor.ClusterStores {
				var old commands.Diffable
				if true { // if object does not exist yet
					old = commands.EmptyDiffer{}
				}
				d, err := differ.Diff(old, c)
				if err != nil {
					return err
				}
				sBuilder.WriteString(d + "\n")
			}

			sBuilder.WriteString("Cluster Stacks\n\n")
			for _, c := range descriptor.GetClusterStacks() {
				var old commands.Diffable
				if true { // if object does not exist yet
					old = importpkg.ClusterStack{
						Name: c.Name,
						BuildImage: importpkg.Source{
							Image: "build-image",
						},
						RunImage: importpkg.Source{
							Image: "run-image",
						},
					}
				}
				d, err := differ.Diff(old, c)
				if err != nil {
					return err
				}
				sBuilder.WriteString(d + "\n")
			}

			sBuilder.WriteString("Cluster Builders\n\n")
			for _, c := range descriptor.GetClusterBuilders() {
				var old commands.Diffable
				if c.Name == "default" {
					old = c
				} else { // if object does not exist yet
					old = commands.EmptyDiffer{}
				}
				d, err := differ.Diff(old, c)
				if err != nil {
					return err
				}
				sBuilder.WriteString(d + "\n")
			}

			ch.Printlnf(sBuilder.String())

			if !force {
				confirmed, err := confirmationProvider.Confirm(confirmMessage)
				if err != nil {
					return err
				}

				if !confirmed {
					return ch.Printlnf("Skipping import")
				}
			}

			if err := importFactory.ImportClusterStores(descriptor.ClusterStores, storeFactory, repository); err != nil {
				return err
			}

			if err := importFactory.ImportClusterStacks(descriptor.GetClusterStacks(), stackFactory); err != nil {
				return err
			}

			if err := importFactory.ImportClusterBuilders(descriptor.GetClusterBuilders(), repository, serviceAccount); err != nil {
				return err
			}

			if err := ch.PrintObjs(importFactory.Objects()); err != nil {
				return err
			}

			return ch.PrintResult("Imported resources created")
		},
	}
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "dependency descriptor filename")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&output, "output", "", "output format. supported formats are: yaml, json")
	cmd.Flags().BoolVar(&force, "force", false, "force import without confirmation")
	commands.SetTLSFlags(cmd, &storeFactory.TLSConfig)
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
