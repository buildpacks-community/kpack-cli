// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
)

func TestExactArgsWithUsage(t *testing.T) {
	cmd := cobra.Command{
		Args: commands.ExactArgsWithUsage(1),
	}
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "some usage")
		return err
	})

	expectedMsg := "accepts 1 arg(s), received 0\n\nsome usage\n"
	err := cmd.ValidateArgs([]string{})

	if err == nil || err.Error() != expectedMsg {
		t.Errorf(`Did not return expected usage error from using wrong number of args.
Expected:
%v
Actual:
%v`, expectedMsg, err)
	}
}

func TestOptionalArgsWithUsage(t *testing.T) {
	cmd := cobra.Command{
		Args: commands.OptionalArgsWithUsage(1),
	}
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "some usage")
		return err
	})

	expectedMsg := "accepts 0 or 1 arg(s), received 2\n\nsome usage\n"
	err := cmd.ValidateArgs([]string{"some-arg-1", "some-arg-2"})

	if err == nil || err.Error() != expectedMsg {
		t.Errorf(`Did not return expected usage error from using wrong number of args.
Expected:
%v
Actual:
%v`, expectedMsg, err)
	}
}
