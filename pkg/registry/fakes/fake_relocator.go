// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Relocator struct {
	skip  bool
	calls []struct {
		Keychain authn.Keychain
		Image    v1.Image
		Dest     string
	}
	writer io.Writer
}

func (r *Relocator) Relocate(keychain authn.Keychain, image v1.Image, dest string) (string, error) {
	r.calls = append(r.calls, struct {
		Keychain authn.Keychain
		Image    v1.Image
		Dest     string
	}{keychain, image, dest})

	digest, err := image.Digest()
	if err != nil {
		return "", err
	}
	sha := digest.String()

	destRef, err := name.ParseReference(dest)
	if err != nil {
		return "", err
	}

	refDigestStr := fmt.Sprintf("%s/%s@%s", destRef.Context().RegistryStr(), destRef.Context().RepositoryStr(), sha)
	var message string
	if r.skip {
		message = fmt.Sprintf("\tSkipping '%s'\n", refDigestStr)
	} else {
		message = fmt.Sprintf("\tUploading '%s'\n", refDigestStr)
	}

	if r.writer == nil {
		r.writer = ioutil.Discard
	}
	_, err = r.writer.Write([]byte(message))
	return refDigestStr, err
}

func (r *Relocator) CallCount() int {
	return len(r.calls)
}

func (r *Relocator) SetSkip(skip bool) {
	r.skip = skip
}

func (r *Relocator) RelocateCall(i int) (authn.Keychain, v1.Image, string) {
	return r.calls[i].Keychain, r.calls[i].Image, r.calls[i].Dest
}