// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func NewPrinter(cmd *cobra.Command) *Logger {
	return &Logger{
		Writer: cmd.OutOrStdout(),
		Err:    cmd.OutOrStderr(),
	}
}

func NewDiscardPrinter() *Logger {
	return &Logger{
		Writer: ioutil.Discard,
		Err:    ioutil.Discard,
	}
}

type Logger struct {
	io.Writer
	Err io.Writer
}

func (l *Logger) Printf(format string, a ...interface{}) {
	l.Write([]byte(fmt.Sprintf(format+"\n", a...)))
}
