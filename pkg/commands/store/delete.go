package store

import (
	"github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewDeleteCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <buildpackage>",
		Short: "Delete buildpackage(s) from store",
		Long: `Deletes existing buildpackage(s) from the buildpack store.

This relies on the image(s) specified to exist in the store and deletes the associated buildpackage
`,
		Example: `tbctl store delete my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271
tbctl store delete my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271 my-registry.com/my-buildpackage/buildpacks_nginx@sha256:eacecf914a57e4f5d6c4459226f6133ecaccb5c462717a09cfeae4763207b9ef
`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			printer := commands.NewPrinter(cmd)

			store, err := cs.KpackClient.ExperimentalV1alpha1().Stores().Get(DefaultStoreName, v1.GetOptions{})
			if err != nil {
				return err
			}

			for _, bpToDelete := range args {
				if !storeContainsBuildpackage(store, bpToDelete) {
					return errors.Errorf("Buildpackage '%s' does not exist in the store", bpToDelete)
				}
			}
			var updatedStoreSources = []v1alpha1.StoreImage{}
			for _, storeImg := range store.Spec.Sources {
				found := false
				for _, bpToDelete := range args {
					if storeImg.Image == bpToDelete {
						found = true
						printer.Printf("Removing buildpackage %s", bpToDelete)
						break
					}
				}
				if !found {
					updatedStoreSources = append(updatedStoreSources, storeImg)
				}
			}

			store.Spec.Sources = updatedStoreSources

			_, err = cs.KpackClient.ExperimentalV1alpha1().Stores().Update(store)
			if err != nil {
				return err
			}

			printer.Printf("Store Updated")
			return nil
		},
	}
	return cmd
}

func storeContainsBuildpackage(store *v1alpha1.Store, buildpackage string) bool {
	for _, source := range store.Spec.Sources {
		if source.Image == buildpackage {
			return true
		}
	}
	return false
}
