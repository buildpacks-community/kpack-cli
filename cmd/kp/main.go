// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/vmware-tanzu/kpack-cli/pkg/rootcommand"
)

func main() {
	log.SetOutput(ioutil.Discard)

	cmd := rootcommand.GetRootCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
