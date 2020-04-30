package stack

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

const (
	StackIdLabel                = "io.buildpacks.stack.id"
	RunImageName                = "run"
	BuildImageName              = "build"
	DefaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
)

type ImageUploader interface {
	Upload(repository, name, image string) (string, v1.Image, error)
}

func NewUpdateCommand(kpackClient versioned.Interface, imageUploader ImageUploader) *cobra.Command {
	var (
		buildImage string
		runImage   string
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a stack",
		Long: `Updates the run and build images of a stack.

The run and build images will be uploaded to the the registry configured on your stack.
Therefore, you must have credentials to access the registry on your machine.
`,
		Example: `tbctl stack update my-stack --build-image my-registry.com/build --run-image my-registry.com/run
tbctl stack update my-stack --build-image ../path/to/build.tar --run-image ../path/to/run.tar
`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := commands.NewPrinter(cmd)

			stack, err := kpackClient.ExperimentalV1alpha1().Stacks().Get(args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}

			repository, ok := stack.Annotations[DefaultRepositoryAnnotation]
			if !ok || repository == "" {
				return errors.Errorf("Unable to find default registry for stack: %s", args[0])
			}

			printer.Printf("Uploading to '%s'...", repository)

			uploadedBuildImageRef, uploadedBuildImage, err := imageUploader.Upload(repository, BuildImageName, buildImage)
			if err != nil {
				return err
			}

			uploadedRunImageRef, uploadedRunImage, err := imageUploader.Upload(repository, RunImageName, runImage)
			if err != nil {
				return err
			}

			uploadedStackId, err := getStackID(uploadedBuildImage, uploadedRunImage)
			if err != nil {
				return err
			}

			if !mutateStack(stack, uploadedBuildImageRef, uploadedRunImageRef, uploadedStackId) {
				printer.Printf("Build and Run images already exist in stack\nStack Unchanged")
				return nil
			}

			_, err = kpackClient.ExperimentalV1alpha1().Stacks().Update(stack)
			if err != nil {
				return err
			}

			printer.Printf("Stack Updated")
			return nil
		},
	}

	cmd.Flags().StringVarP(&buildImage, "build-image", "b", "", "build image tag or local tar file path")
	cmd.Flags().StringVarP(&runImage, "run-image", "r", "", "run image tag or local tar file path")
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")

	return cmd
}

func getStackID(buildImg, runImg v1.Image) (string, error) {
	buildStack, err := getStackLabel(buildImg, StackIdLabel)
	if err != nil {
		return "", err
	}

	runStack, err := getStackLabel(runImg, StackIdLabel)
	if err != nil {
		return "", err
	}

	if buildStack != runStack {
		return "", errors.Errorf("build stack '%s' does not match run stack '%s'", buildStack, runStack)
	}

	return buildStack, nil
}

func getStackLabel(image v1.Image, label string) (string, error) {
	config, err := image.ConfigFile()
	if err != nil {
		return "", err
	}
	labels := config.Config.Labels
	id, ok := labels[label]
	if !ok {
		return "", errors.New("invalid stack image")
	}
	return id, nil
}

func mutateStack(stack *expv1alpha1.Stack, buildImageRef, runImageRef, stackId string) bool {
	if stack.Status.BuildImage.LatestImage != buildImageRef && stack.Status.RunImage.LatestImage != runImageRef {
		stack.Spec.BuildImage.Image = buildImageRef
		stack.Spec.RunImage.Image = runImageRef
		stack.Spec.Id = stackId
		return true
	}
	return false
}
