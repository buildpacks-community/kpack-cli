// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack_test

import (
	"testing"

	"github.com/sclevine/spec"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/buildpack"
)

func TestBuildpackSaveCommand(t *testing.T) {
	spec.Run(t, "TestBuildpackSaveCommandCreate", testCreateCommand(buildpack.NewSaveCommand))
	spec.Run(t, "TestBuildpackSaveCommandPatch", testPatchCommand(buildpack.NewSaveCommand))
}
