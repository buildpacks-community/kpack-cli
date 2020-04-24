package commands

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
)

type TableWriter struct {
	numColumns int
	writer     *tabwriter.Writer
}

func NewTableWriter(out io.Writer, headers ...string) (*TableWriter, error) {
	writer := tabwriter.NewWriter(out, 0, 4, 4, ' ', 0)

	_, err := fmt.Fprintln(writer, strings.ToUpper(strings.Join(headers, "\t")))
	if err != nil {
		return nil, err
	}

	return &TableWriter{
		numColumns: len(headers),
		writer:     writer,
	}, nil
}

func (w *TableWriter) AddRow(columns ...string) error {
	if len(columns) != w.numColumns {
		return errors.New("incorrect number of columns for row")
	}

	_, err := fmt.Fprintln(w.writer, strings.Join(columns, "\t"))
	return err
}

func (w *TableWriter) Write() error {
	_, err := fmt.Fprintln(w.writer, "")
	if err != nil {
		return err
	}
	return w.writer.Flush()
}
