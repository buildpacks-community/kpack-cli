// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
	"github.com/vmware-tanzu/kpack-cli/pkg/stackimage"
)

type Uploader interface {
	UploadStackImages(keychain authn.Keychain, buildImageTag, runImageTag, dest string) (string, string, error)
	ValidateStackIDs(keychain authn.Keychain, buildImageTag, runImageTag string) (string, error)
	UploadedBuildImageRef(keychain authn.Keychain, imageTag, dest string) (string, error)
	UploadedRunImageRef(keychain authn.Keychain, imageTag, dest string) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	PrintStatus(format string, args ...interface{}) error
}

type Factory struct {
	Uploader Uploader
	Printer  Printer
}

func NewFactory(printer Printer, relocator registry.Relocator, fetcher registry.Fetcher) *Factory {
	return &Factory{
		Uploader: &stackimage.Uploader{
			Fetcher:   fetcher,
			Relocator: relocator,
		},
		Printer: printer,
	}
}

func (f *Factory) MakeStack(keychain authn.Keychain, name, buildImageTag, runImageTag string, kpConfig config.KpConfig) (*v1alpha1.ClusterStack, error) {
	stackID, err := f.validate(keychain, buildImageTag, runImageTag)
	if err != nil {
		return nil, err
	}

	canonicalRepo, err := kpConfig.CanonicalRepository()
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", canonicalRepo); err != nil {
		return nil, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, err := f.Uploader.UploadStackImages(keychain, buildImageTag, runImageTag, canonicalRepo)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: stackID,
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedBuildImageRef,
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedRunImageRef,
			},
		},
	}, nil
}

func (f *Factory) UpdateStack(keychain authn.Keychain, stack *v1alpha1.ClusterStack, buildImageTag, runImageTag string, kpConfig config.KpConfig) (bool, error) {
	stackID, err := f.validate(keychain, buildImageTag, runImageTag)
	if err != nil {
		return false, err
	}

	canonicalRepo, err := kpConfig.CanonicalRepository()
	if err != nil {
		return false, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", canonicalRepo); err != nil {
		return false, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, err := f.Uploader.UploadStackImages(keychain, buildImageTag, runImageTag, canonicalRepo)
	if err != nil {
		return false, err
	}

	if wasUpdated, err := wasUpdated(stack, relocatedBuildImageRef, relocatedRunImageRef, stackID); err != nil {
		return false, err
	} else if !wasUpdated {
		return false, f.Printer.Printlnf("Build and Run images already exist in stack")
	}
	return true, nil
}

func (f *Factory) RelocatedBuildImage(keychain authn.Keychain, kpConfig config.KpConfig, tag string) (string, error) {
	canonicalRepo, err := kpConfig.CanonicalRepository()
	if err != nil {
		return "", err
	}

	return f.Uploader.UploadedBuildImageRef(keychain, tag, canonicalRepo)
}

func (f *Factory) RelocatedRunImage(keychain authn.Keychain, kpConfig config.KpConfig, tag string) (string, error) {
	canonicalRepo, err := kpConfig.CanonicalRepository()
	if err != nil {
		return "",err
	}

	return f.Uploader.UploadedRunImageRef(keychain, tag, canonicalRepo)
}

func (f *Factory) validate(keychain authn.Keychain, buildTag, runTag string) (string, error) {
	return f.Uploader.ValidateStackIDs(keychain, buildTag, runTag)
}

func wasUpdated(stack *v1alpha1.ClusterStack, buildImageRef, runImageRef, stackId string) (bool, error) {
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
		return "", errors.Errorf("failed to get image digest from reference %q", ref)
	}
	return s[1], nil
}
