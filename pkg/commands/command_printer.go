// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import "C"
import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
)

type CommandPrinter struct {
	dryRun          bool
	outputResource  bool
	outWriter       io.Writer
	errWriter       io.Writer
	builder         strings.Builder
	resourcePrinter ResourcePrinter
}

func NewCommandPrinter(cmd *cobra.Command) (*CommandPrinter, error) {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return nil, err
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}

	var resourcePrinter ResourcePrinter

	outputResource := len(output) > 0
	if outputResource {
		resourcePrinter, err = NewResourcePrinter(output)
		if err != nil {
			return nil, err
		}
	}

	return &CommandPrinter{
		dryRun,
		outputResource,
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		strings.Builder{},
		resourcePrinter,
	}, nil
}

func (cp CommandPrinter) PrintObj(obj runtime.Object) error {
	if cp.outputResource {
		return cp.resourcePrinter.PrintObject(obj, cp.outWriter)
	}
	return nil
}

func (cp CommandPrinter) PrintResult(format string, a ...interface{}) error {
	cp.builder.Reset()

	str := fmt.Sprintf(format, a...)
	_, err := cp.builder.WriteString(str)
	if err != nil {
		return err
	}

	if cp.dryRun {
		_, err = cp.builder.WriteString(" (dry run)")
		if err != nil {
			return err
		}
	}
	cp.builder.WriteString("\n")

	_, err = cp.ResultWriter().Write([]byte(cp.builder.String()))
	return err
}

func (cp CommandPrinter) Printlnf(format string, a ...interface{}) error {
	return cp.printf(format+"\n", a...)
}

func (cp CommandPrinter) printf(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(cp.TextWriter(), format, a...)
	return err
}

func (cp CommandPrinter) IsDryRun() bool {
	return cp.dryRun
}

func (cp CommandPrinter) TextWriter() io.Writer {
	if cp.outputResource {
		return cp.errWriter
	} else {
		return cp.outWriter
	}
}

func (cp CommandPrinter) ResultWriter() io.Writer {
	if cp.outputResource {
		return ioutil.Discard
	} else {
		return cp.outWriter
	}
}
