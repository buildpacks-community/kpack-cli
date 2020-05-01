package stack

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Factory struct {
	Uploader    ImageUploader
	BuildImage  string
	RunImage    string
	DefaultRepo string
}

const (
	DefaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
	StackIdLabel                = "io.buildpacks.stack.id"
	RunImageName                = "run"
	BuildImageName              = "build"
)

func (f *Factory) MakeStack(name string) (*expv1alpha1.Stack, error) {
	buildImg, err := f.Uploader.Read(f.BuildImage)
	if err != nil {
		return nil, err
	}

	runImg, err := f.Uploader.Read(f.RunImage)
	if err != nil {
		return nil, nil
	}

	stackId, err := GetStackID(buildImg, runImg)
	if err != nil {
		return nil, err
	}

	buildImageRef, err := f.Uploader.Upload(buildImg, f.DefaultRepo, BuildImageName)
	if err != nil {
		return nil, err
	}

	runImageRef, err := f.Uploader.Upload(runImg, f.DefaultRepo, RunImageName)
	if err != nil {
		return nil, err
	}

	return &expv1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				DefaultRepositoryAnnotation: f.DefaultRepo,
			},
		},

		Spec: expv1alpha1.StackSpec{
			Id: stackId,
			BuildImage: expv1alpha1.StackSpecImage{
				Image: buildImageRef,
			},
			RunImage: expv1alpha1.StackSpecImage{
				Image: runImageRef,
			},
		},
	}, nil

}

func GetStackID(buildImg, runImg v1.Image) (string, error) {
	buildStack, err := imagehelpers.GetStringLabel(buildImg, StackIdLabel)
	if err != nil {
		return "", err
	}

	runStack, err := imagehelpers.GetStringLabel(runImg, StackIdLabel)
	if err != nil {
		return "", err
	}

	if buildStack != runStack {
		return "", errors.Errorf("build stack '%s' does not match run stack '%s'", buildStack, runStack)
	}

	return buildStack, nil
}
