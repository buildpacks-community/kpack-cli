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

type CommandHelper struct {
	dryRun         bool
	outputResource bool
	wait           bool

	outWriter       io.Writer
	errWriter       io.Writer

	builder         strings.Builder
	resourcePrinter ResourcePrinter
}

func NewCommandHelper(cmd *cobra.Command) (*CommandHelper, error) {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return nil, err
	}

	output := ""
	if cmd.Flags().Changed("output") {
		output, err = cmd.Flags().GetString("output")
		if err != nil {
			return nil, err
		}
	}

	wait := false
	flag := cmd.Flags().Lookup("wait")
	if flag != nil {
		wait, err = cmd.Flags().GetBool("wait")
		if err != nil {
			return nil, err
		}
	}

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

	return &CommandHelper{
		dryRun,
		outputResource,
		wait,
		cmd.OutOrStdout(),
		cmd.ErrOrStderr(),
		strings.Builder{},
		resourcePrinter,
	}, nil
}

func (cp CommandHelper) IsDryRun() bool {
	return cp.dryRun
}

func (cp CommandHelper) CanWait() bool {
	return cp.wait && !cp.dryRun && !cp.outputResource

}

func (cp CommandHelper) PrintObj(obj runtime.Object) error {
	if cp.outputResource {
		return cp.resourcePrinter.PrintObject(obj, cp.outWriter)
	}
	return nil
}

func (cp CommandHelper) PrintResult(format string, a ...interface{}) error {
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

	_, err = cp.OutOrDiscardWriter().Write([]byte(cp.builder.String()))
	return err
}

func (cp CommandHelper) Printlnf(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(cp.OutOrErrWriter(), format+"\n", a...)
	return err
}

func (cp CommandHelper) OutOrErrWriter() io.Writer {
	if cp.outputResource {
		return cp.errWriter
	} else {
		return cp.outWriter
	}
}

func (cp CommandHelper) OutOrDiscardWriter() io.Writer {
	if cp.outputResource {
		return ioutil.Discard
	} else {
		return cp.outWriter
	}
}
