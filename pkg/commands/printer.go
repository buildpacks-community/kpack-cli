// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func NewPrinter(cmd *cobra.Command) *Logger {
	return &Logger{
		Out: cmd.OutOrStdout(),
		Err: cmd.OutOrStderr(),
	}
}

type Logger struct {
	Out io.Writer
	Err io.Writer
}

func (l *Logger) Printf(format string, a ...interface{}) {
	l.Out.Write([]byte(fmt.Sprintf(format+"\n", a...)))
}
