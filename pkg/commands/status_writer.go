// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/pkg/errors"
)

type StatusWriter struct {
	writer *tabwriter.Writer
}

const StatusWriterTabWidth = 4
const StatusWriterPadding = 4

func NewStatusWriter(out io.Writer) *StatusWriter {
	return &StatusWriter{
		writer: tabwriter.NewWriter(out, 0, StatusWriterTabWidth, StatusWriterPadding, ' ', 0),
	}
}

func (s *StatusWriter) AddBlock(header string, items ...string) error {
	if len(items)%2 != 0 {
		return errors.Errorf("block must contain an equal number of items")
	}

	if header != "" {
		_, err := fmt.Fprintln(s.writer, header)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(items); i += 2 {
		value := items[i+1]
		if value == "" {
			value = "--"
		}

		_, err := fmt.Fprintf(s.writer, "%s:\t%s\n", items[i], value)
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(s.writer, "")
	return err
}

func (s *StatusWriter) Write() error {
	return s.writer.Flush()
}
