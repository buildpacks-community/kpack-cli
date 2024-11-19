// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/sclevine/spec"

	cbcmds "github.com/buildpacks-community/kpack-cli/pkg/commands/clusterbuilder"
)

func TestClusterBuilderSaveCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderSaveCommandCreate", testCreateCommand(cbcmds.NewSaveCommand))
	spec.Run(t, "TestClusterBuilderSaveCommandPatch", testPatchCommand(cbcmds.NewSaveCommand))
}
