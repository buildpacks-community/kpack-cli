package store

import (
	"encoding/json"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

const (
	defaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
	kubectlLastAppliedConfig    = "kubectl.kubernetes.io/last-applied-configuration"
)

type BuildpackageUploader interface {
	Upload(repository, buildPackage string) (string, error)
}

type Factory struct {
	Uploader          BuildpackageUploader
	Buildpackages     []string
	DefaultRepository string
	Printer           *commands.Logger
}

func (f *Factory) MakeStore(name string) (*v1alpha1.Store, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}

	newStore := &v1alpha1.Store{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.StoreKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				defaultRepositoryAnnotation: f.DefaultRepository,
			},
		},
		Spec: v1alpha1.StoreSpec{},
	}

	f.Printer.Printf("Uploading to '%s'...", f.DefaultRepository)

	var uploaded []string
	for _, buildpackage := range f.Buildpackages {
		uploadedBp, err := f.Uploader.Upload(f.DefaultRepository, buildpackage)
		if err != nil {
			return nil, err
		}

		newStore.Spec.Sources = append(newStore.Spec.Sources, v1alpha1.StoreImage{
			Image: uploadedBp,
		})

		uploaded = append(uploaded, uploadedBp)
	}

	marshal, err := json.Marshal(newStore)
	if err != nil {
		return nil, err
	}

	newStore.Annotations[kubectlLastAppliedConfig] = string(marshal)

	return newStore, nil
}

func (f *Factory) validate() error {
	if len(f.Buildpackages) < 1 {
		return errors.New("At least one buildpackage must be provided")
	}

	_, err := name.ParseReference(f.DefaultRepository, name.WeakValidation)

	return err
}
