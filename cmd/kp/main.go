// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io"
	"log"
	"os"

	"github.com/buildpacks-community/kpack-cli/pkg/rootcommand"
)

func main() {
	log.SetOutput(io.Discard)

	cmd := rootcommand.GetRootCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
