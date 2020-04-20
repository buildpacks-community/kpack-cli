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

func NewStatusWriter(out io.Writer) *StatusWriter {
	return &StatusWriter{
		writer: tabwriter.NewWriter(out, 0, 4, 4, ' ', 0),
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
		_, err := fmt.Fprintf(s.writer, "%s:\t%s\n", items[i], items[i+1])
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
