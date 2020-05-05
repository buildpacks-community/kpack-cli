package image

import (
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
)

type Relocator struct {
}

func (*Relocator) Relocate(image v1.Image, dest string) (string, error) {
	ref, err := name.ParseReference(dest, name.WeakValidation)
	if err != nil {
		return "", errors.WithStack(err)
	}

	refName := fmt.Sprintf("%s/%s", ref.Context().RegistryStr(), ref.Context().RepositoryStr())
	ref, err = name.ParseReference(refName, name.WeakValidation)
	if err != nil {
		return "", errors.WithStack(err)
	}

	err = remote.Write(ref, image, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", newImageAccessError(refName, err)
	}

	digest, err := image.Digest()
	if err != nil {
		return "", errors.WithStack(err)
	}

	return fmt.Sprintf("%s@%s", refName, digest.String()), remote.Tag(ref.Context().Tag(timestampTag()), image, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}

func timestampTag() string {
	now := time.Now()
	return fmt.Sprintf("%s%02d%02d%02d", now.Format("20060102"), now.Hour(), now.Minute(), now.Second())
}
