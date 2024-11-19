// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/sclevine/spec"

	imgcmds "github.com/buildpacks-community/kpack-cli/pkg/commands/image"
)

func TestImageSaveCommand(t *testing.T) {
	spec.Run(t, "TestImageSaveCommandCreate", testCreateCommand(imgcmds.NewSaveCommand))
	spec.Run(t, "TestImageSaveCommandPatch", testPatchCommand(imgcmds.NewSaveCommand))
}
