// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle_test

import (
	"testing"

	"github.com/sclevine/spec"

	clusterstackcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstack"
)

func TestSaveCommand(t *testing.T) {
	spec.Run(t, "TestSaveCommandCreate", testCreateCommand(clusterstackcmds.NewSaveCommand))
	spec.Run(t, "TestSaveCommandUpdate", testUpdateCommand(clusterstackcmds.NewSaveCommand))
}
