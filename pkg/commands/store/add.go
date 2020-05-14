package store

import (
	"strings"

	"github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

const (
	defaultStoreName            = "default"
	defaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
)

type BuildpackageUploader interface {
	Upload(repository, buildPackage string) (string, error)
}

func NewAddCommand(kpackClient versioned.Interface, buildpackUploader BuildpackageUploader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <buildpackage>",
		Short: "Create an image configuration",
		Long: `Upload builpackage(s) to a the buildpack store.

Buildpackages will be uploaded to the the registry configured on your store.
Therefore, you must have credentials to access the registry on your machine.
`,
		Example: `tbctl store add my-registry.com/my-buildpackage
tbctl store add my-registry.com/my-buildpackage my-registry.com/my-other-buildpackage my-registry.com/my-third-buildpackage
tbctl store add ../path/to/my-local-buildpackage.cnb
`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := commands.NewPrinter(cmd)

			store, err := kpackClient.ExperimentalV1alpha1().Stores().Get(defaultStoreName, v1.GetOptions{})
			if err != nil {
				return err
			}

			repository, ok := store.Annotations[defaultRepositoryAnnotation]
			if !ok || repository == "" {
				return errors.Errorf("Unable to find default registry for store: %s", defaultStoreName)
			}

			printer.Printf("Uploading to '%s'...", repository)

			var uploaded []string
			for _, buildpackage := range args {
				uploadedBp, err := buildpackUploader.Upload(repository, buildpackage)
				if err != nil {
					return err
				}
				uploaded = append(uploaded, uploadedBp)
			}

			storeUpdated := false
			for _, uploadedBp := range uploaded {
				if storeContains(store, uploadedBp) {
					printer.Printf("Buildpackage '%s' already exists in the store", uploadedBp)
					continue
				}

				store.Spec.Sources = append(store.Spec.Sources, v1alpha1.StoreImage{
					Image: uploadedBp,
				})
				storeUpdated = true
				printer.Printf("Added Buildpackage '%s'", uploadedBp)
			}

			if !storeUpdated {
				printer.Printf("Store Unchanged")
				return nil
			}

			_, err = kpackClient.ExperimentalV1alpha1().Stores().Update(store)
			if err != nil {
				return err
			}

			printer.Printf("Store Updated")
			return nil
		},
	}
	return cmd
}

func storeContains(store *v1alpha1.Store, buildpackage string) bool {
	digest := strings.Split(buildpackage, "@")[1]

	for _, image := range store.Spec.Sources {
		parts := strings.Split(image.Image, "@")
		if len(parts) != 2 {
			continue
		}

		if parts[1] == digest {
			return true
		}
	}
	return false
}
