// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	"github.com/sclevine/spec"

	buildercmds "github.com/buildpacks-community/kpack-cli/pkg/commands/builder"
)

func TestBuilderSaveCommand(t *testing.T) {
	spec.Run(t, "TestBuilderSaveCommandCreate", testCreateCommand(buildercmds.NewSaveCommand))
	spec.Run(t, "TestBuilderSaveCommandPatch", testPatchCommand(buildercmds.NewSaveCommand))
}
