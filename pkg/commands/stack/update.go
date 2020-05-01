package stack

import (
	"strings"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	pkgstack "github.com/pivotal/build-service-cli/pkg/stack"
)

const (
	RunImageName                = "run"
	BuildImageName              = "build"
	DefaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
)

func NewUpdateCommand(kpackClient versioned.Interface, imageUploader pkgstack.ImageUploader) *cobra.Command {
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

			buildImg, err := imageUploader.Read(buildImage)
			if err != nil {
				return err
			}

			runImg, err := imageUploader.Read(runImage)
			if err != nil {
				return err
			}

			uploadedStackId, err := pkgstack.GetStackID(buildImg, runImg)
			if err != nil {
				return err
			}

			uploadedBuildImageRef, err := imageUploader.Upload(buildImg, repository, BuildImageName)
			if err != nil {
				return err
			}

			uploadedRunImageRef, err := imageUploader.Upload(runImg, repository, RunImageName)
			if err != nil {
				return err
			}

			if mutated, err := mutateStack(stack, uploadedBuildImageRef, uploadedRunImageRef, uploadedStackId); err != nil {
				return err
			} else if !mutated {
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

func mutateStack(stack *expv1alpha1.Stack, buildImageRef, runImageRef, stackId string) (bool, error) {
	oldBuildDigest, err := getDigest(stack.Status.BuildImage.LatestImage)
	if err != nil {
		return false, err
	}

	newBuildDigest, err := getDigest(buildImageRef)
	if err != nil {
		return false, err
	}

	oldRunDigest, err := getDigest(stack.Status.RunImage.LatestImage)
	if err != nil {
		return false, err
	}

	newRunDigest, err := getDigest(runImageRef)
	if err != nil {
		return false, err
	}

	if oldBuildDigest != newBuildDigest && oldRunDigest != newRunDigest {
		stack.Spec.BuildImage.Image = buildImageRef
		stack.Spec.RunImage.Image = runImageRef
		stack.Spec.Id = stackId
		return true, nil
	}
	return false, nil
}

func getDigest(ref string) (string, error) {
	s := strings.Split(ref, "@")
	if len(s) != 2 {
		return "", errors.New("failed to get image digest")
	}
	return s[1], nil
}
