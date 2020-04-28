package stack

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

const (
	stackID                     = "io.buildpacks.stack.id"
	runImageName                = "run"
	buildImageName              = "build"
	defaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
)

type ImageUploader interface {
	Upload(repository, image, name string) (string, v1.Image, error)
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

			repository, ok := stack.Annotations[defaultRepositoryAnnotation]
			if !ok || repository == "" {
				return errors.Errorf("Unable to find default registry for stack: %s", args[0])
			}

			printer.Printf("Uploading to '%s'...", repository)

			uploadedBuildImageRef, uploadedBuildImage, err := imageUploader.Upload(repository, buildImage)
			if err != nil {
				return err
			}

			uploadedRunImageRef, uploadedRunImage, err := imageUploader.Upload(repository, runImage)
			if err != nil {
				return err
			}

			uploadedStackId, err := getStackID(uploadedBuildImage, uploadedRunImage)
			if err != nil {
				return err
			}

			stackUpdated := false

			if stack.Status.BuildImage.LatestImage != uploadedBuildImageRef {
				stack.Spec.BuildImage.Image = uploadedBuildImageRef
				stackUpdated = true
			}

			if stack.Status.RunImage.LatestImage != uploadedRunImageRef {
				stack.Spec.RunImage.Image = uploadedRunImageRef
				stackUpdated = true
			}

			if stack.Status.Id != uploadedStackId {
				stack.Spec.Id = uploadedStackId
				stackUpdated = true
			}

			if !stackUpdated {
				printer.Printf("Stack Unchanged")
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
	cmd.Flags().StringVarP(&buildImage, "run-image", "r", "", "run image tag or local tar file path")
	_ = cmd.MarkFlagRequired("build-image")
	_ = cmd.MarkFlagRequired("run-image")

	return cmd
}

func getStackID(buildImg, runImg v1.Image) (string, error) {
	buildStack, err := getStackLabel(buildImg, stackID)
	if err != nil {
		return "", err
	}

	runStack, err := getStackLabel(runImg, stackID)
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
