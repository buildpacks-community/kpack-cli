// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

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
	dryRun bool
	output bool
	wait   bool

	outWriter io.Writer
	errWriter io.Writer

	objPrinter k8s.ObjectPrinter
	strBuilder strings.Builder
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

	var objPrinter k8s.ObjectPrinter

	outputResource := len(output) > 0
	if outputResource {
		objPrinter, err = k8s.NewObjectPrinter(output)
		if err != nil {
			return nil, err
		}
	}

	return &CommandHelper{
		dryRun:     dryRun,
		output:     outputResource,
		wait:       wait,
		outWriter:  cmd.OutOrStdout(),
		errWriter:  cmd.ErrOrStderr(),
		objPrinter: objPrinter,
		strBuilder: strings.Builder{},
	}, nil
}

func (ch CommandHelper) IsDryRun() bool {
	return ch.dryRun
}

func (ch CommandHelper) ShouldWait() bool {
	return ch.wait && !ch.dryRun && !ch.output
}

func (ch CommandHelper) PrintObjs(objs []runtime.Object) error {
	if ch.output {
		for _, obj := range objs {
			if err := ch.objPrinter.PrintObject(obj, ch.outWriter); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ch CommandHelper) PrintObj(obj runtime.Object) error {
	if ch.output {
		return ch.objPrinter.PrintObject(obj, ch.outWriter)
	}
	return nil
}

func (ch CommandHelper) PrintResult(format string, a ...interface{}) error {
	ch.strBuilder.Reset()

	str := fmt.Sprintf(format, a...)
	_, err := ch.strBuilder.WriteString(str)
	if err != nil {
		return err
	}

	if ch.dryRun {
		_, err = ch.strBuilder.WriteString(" (dry run)")
		if err != nil {
			return err
		}
	}
	ch.strBuilder.WriteString("\n")

	_, err = ch.OutOrDiscardWriter().Write([]byte(ch.strBuilder.String()))
	return err
}

func (ch CommandHelper) Printlnf(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(ch.OutOrErrWriter(), format+"\n", a...)
	return err
}

func (ch CommandHelper) OutOrErrWriter() io.Writer {
	if ch.output {
		return ch.errWriter
	} else {
		return ch.outWriter
	}
}

func (ch CommandHelper) OutOrDiscardWriter() io.Writer {
	if ch.output {
		return ioutil.Discard
	} else {
		return ch.outWriter
	}
}

func (ch CommandHelper) Writer() io.Writer {
	return ch.OutOrErrWriter()
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
