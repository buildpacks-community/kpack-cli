// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/sclevine/spec"
	storecmds "github.com/buildpacks-community/kpack-cli/pkg/commands/clusterstore"
)

func TestClusterStoreSaveCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreSaveCommandCreate", testCreateCommand(storecmds.NewSaveCommand))
	spec.Run(t, "TestClusterStoreSaveCommandUpdate", testAddCommand(storecmds.NewSaveCommand))
}
