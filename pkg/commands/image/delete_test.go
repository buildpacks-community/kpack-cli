package image_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
)

func TestImageDeleteCommand(t *testing.T) {
	spec.Run(t, "TestImageDeleteCommand", testImageDeleteCommand)
}

func testImageDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		out          = &bytes.Buffer{}
		imageDeleter = newFakeImageDeleter()
		deleteCmd    = &image.DeleteCommand{
			Out:              out,
			Deleter:          imageDeleter,
			DefaultNamespace: "default-namespace",
		}
	)

	when("a namespace is provided", func() {
		it("deletes the image", func() {
			err := deleteCmd.Execute("test-namespace", "test-image-1")
			require.NoError(t, err)
			require.Equal(t, out.String(), "test-image-1 deleted\n")
			require.Contains(t, imageDeleter.deletedImages, "test-namespace")
			require.Contains(t, imageDeleter.deletedImages["test-namespace"], "test-image-1")
		})
	})

	when("a namespace is not provided", func() {
		it("deletes the image using the default namespace", func() {
			err := deleteCmd.Execute("", "test-image-1")
			require.NoError(t, err)
			require.Equal(t, out.String(), "test-image-1 deleted\n")
			require.Contains(t, imageDeleter.deletedImages, "default-namespace")
			require.Contains(t, imageDeleter.deletedImages["default-namespace"], "test-image-1")
		})
	})

	when("the deleter returns an error", func() {
		imageDeleter.err = errors.New("some deleter error")

		it("bubbles up the error", func() {
			err := deleteCmd.Execute("test-namespace", "test-image-1")
			require.Error(t, err, "some deleter error")
		})
	})
}

type fakeImageDeleter struct {
	deletedImages map[string][]string
	err           error
}

func newFakeImageDeleter() *fakeImageDeleter {
	return &fakeImageDeleter{
		deletedImages: map[string][]string{},
	}
}

func (f *fakeImageDeleter) Delete(namespace, name string) error {
	if f.err != nil {
		return f.err
	}

	if names, ok := f.deletedImages[namespace]; ok {
		f.deletedImages[namespace] = append(names, name)
	} else {
		f.deletedImages[namespace] = []string{name}
	}

	return nil
}
