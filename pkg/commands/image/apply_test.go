package image_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
)

func TestImageApplyCommand(t *testing.T) {
	spec.Run(t, "TestImageApplyCommand", testImageApplyCommand)
}

func testImageApplyCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		out          = &bytes.Buffer{}
		imageApplier = &fakeImageApplier{}
		applyCmd     = &image.ApplyCommand{
			Out:     out,
			Applier: imageApplier,
		}
	)

	when("a valid image config exists", func() {
		it("returns a success message from the image applier", func() {
			err := applyCmd.Execute("./testdata/image.yaml")
			require.NoError(t, err)

			require.Equal(t, "test-image created\n", out.String())
		})
	})

	when("the image config is invalid", func() {
		imageApplier.err = errors.New("some applier error")

		it("returns an error message from the image applier", func() {
			err := applyCmd.Execute("./testdata/image.yaml")
			require.Error(t, err, "some applier error")
		})
	})

	when("a valid image config does not exist", func() {
		it("returns an error message", func() {
			err := applyCmd.Execute("does-not-exist")
			require.Error(t, err, `the path "does-not-exist" does not exist`)
		})
	})
}

type fakeImageApplier struct {
	err error
}

func (f *fakeImageApplier) Apply(_ *v1alpha1.Image) error {
	return f.err
}
