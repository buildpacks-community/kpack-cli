// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterbuildpack"
)

func TestClusterBuildpackSaveCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackSaveCommandCreate", testCreateCommand(clusterbuildpack.NewSaveCommand))
	spec.Run(t, "TestClusterBuildpackSaveCommandPatch", testPatchCommand(clusterbuildpack.NewSaveCommand))
}
