// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ghodss/yaml"
	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	importpkg "github.com/pivotal/build-service-cli/pkg/import"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

func NewImportCommand(
	clientSetProvider k8s.ClientSetProvider,
	bpUploader clusterstore.BuildpackageUploader,
	stackUploader clusterstack.Uploader,
	differ importpkg.Differ,
	timestampProvider TimestampProvider,
	confirmationProvider ConfirmationProvider) *cobra.Command {

	var (
		filename  string
		force     bool
		tlsConfig registry.TLSConfig
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

			storeFactory := &clusterstore.Factory{
				Uploader:     bpUploader,
				TLSConfig:    tlsConfig,
				Repository:   repository,
				Printer:      ch,
				ValidateOnly: ch.ValidateOnly(),
			}

			stackFactory := &clusterstack.Factory{
				Uploader:     stackUploader,
				TLSConfig:    tlsConfig,
				Repository:   repository,
				Printer:      ch,
				ValidateOnly: ch.ValidateOnly(),
			}

			importDiffer := &importpkg.ImportDiffer{
				Differ:         differ,
				StoreRefGetter: storeFactory,
				StackRefGetter: stackFactory,
			}

			importer := importer{
				client:            cs.KpackClient,
				commandHelper:     ch,
				timestampProvider: timestampProvider,
			}

			hasChanges, err := showChanges(descriptor, importDiffer, cs.KpackClient, ch)
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

			if err := importer.importClusterStores(descriptor.ClusterStores, storeFactory); err != nil {
				return err
			}

			if err := importer.importClusterStacks(descriptor.GetClusterStacks(), stackFactory); err != nil {
				return err
			}

			if err := importer.importClusterBuilders(descriptor.GetClusterBuilders(), repository, serviceAccount); err != nil {
				return err
			}

			if err := ch.PrintObjs(importer.objects()); err != nil {
				return err
			}

			return ch.PrintResult("Imported resources")
		},
	}
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "dependency descriptor filename")
	cmd.Flags().BoolVar(&force, "force", false, "force import without confirmation")
	commands.SetDryRunOutputFlags(cmd)
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

func showChanges(descriptor importpkg.DependencyDescriptor, importDiffer *importpkg.ImportDiffer, kClient kpack.Interface, ch *commands.CommandHelper) (hasChanges bool, err error) {
	hasChanges = false
	var changes strings.Builder
	changes.WriteString("Changes\n\n")
	changes.WriteString("ClusterStores\n\n")

	var curDiff strings.Builder
	for _, cs := range descriptor.ClusterStores {
		curStore, err := kClient.KpackV1alpha1().ClusterStores().Get(cs.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return false, err
		}
		if k8serrors.IsNotFound(err) {
			curStore = nil
		}

		cStoreDiff, err := importDiffer.DiffClusterStore(curStore, cs)
		if err != nil {
			return false, err
		}
		if cStoreDiff != "" {
			hasChanges = true
			curDiff.WriteString(cStoreDiff + "\n\n")
		}
	}
	if curDiff.String() == "" {
		curDiff.WriteString("No Changes\n\n")
	}
	changes.WriteString(curDiff.String())

	changes.WriteString("ClusterStacks\n\n")
	curDiff.Reset()
	for _, cs := range descriptor.GetClusterStacks() {
		curStack, err := kClient.KpackV1alpha1().ClusterStacks().Get(cs.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return false, err
		}
		if k8serrors.IsNotFound(err) {
			curStack = nil
		}

		cStackDiff, err := importDiffer.DiffClusterStack(curStack, cs)
		if err != nil {
			return false, err
		}
		if cStackDiff != "" {
			hasChanges = true
			curDiff.WriteString(cStackDiff + "\n\n")
		}
	}
	if curDiff.String() == "" {
		curDiff.WriteString("No Changes\n\n")
	}
	changes.WriteString(curDiff.String())

	changes.WriteString("ClusterBuilders\n\n")
	curDiff.Reset()
	for _, cb := range descriptor.GetClusterBuilders() {
		curBuilder, err := kClient.KpackV1alpha1().ClusterBuilders().Get(cb.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return false, err
		}
		if k8serrors.IsNotFound(err) {
			curBuilder = nil
		}

		cBuilderDiff, err := importDiffer.DiffClusterBuilder(curBuilder, cb)
		if err != nil {
			return false, err
		}
		if cBuilderDiff != "" {
			hasChanges = true
			curDiff.WriteString(cBuilderDiff + "\n\n")
		}
	}
	if curDiff.String() == "" {
		curDiff.WriteString("No Changes\n\n")
	}
	changes.WriteString(curDiff.String())

	return hasChanges, ch.Printlnf(changes.String())
}
