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

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

type CommandHelper struct {
	dryRun         bool
	outputResource bool
	wait           bool

	outWriter io.Writer
	errWriter io.Writer

	objectPrinter k8s.ResourcePrinter
	builder       strings.Builder
}

func NewCommandHelper(cmd *cobra.Command) (*CommandHelper, error) {
	dryRun, err := getBoolFlag("dry-run", cmd)
	if err != nil {
		return nil, err
	}

	output, err := getStringFlag("output", cmd)
	if err != nil {
		return nil, err
	}

	wait, err := getBoolFlag("wait", cmd)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	var resourcePrinter k8s.ResourcePrinter

	outputResource := len(output) > 0
	if outputResource {
		resourcePrinter, err = k8s.NewObjectPrinter(output)
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
		resourcePrinter,
		strings.Builder{},
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
		return cp.objectPrinter.PrintObject(obj, cp.outWriter)
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

func (cp CommandHelper) Writer() io.Writer {
	return cp.OutOrErrWriter()
}

func getBoolFlag(name string, cmd *cobra.Command) (bool, error) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return false, nil
	}

	if !cmd.Flags().Changed(name) {
		return false, nil
	}

	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		return value, err
	}
	return value, nil
}

func getStringFlag(name string, cmd *cobra.Command) (string, error) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return "", nil
	}

	if !cmd.Flags().Changed(name) {
		return "", nil
	}

	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return value, err
	}
	return value, nil
}

