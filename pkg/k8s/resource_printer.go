// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	FormatYAML string = "yaml"
	FormatJSON string = "json"
)

type ObjectPrinter interface {
	PrintObject(obj runtime.Object, w io.Writer) error
}

func NewObjectPrinter(format string) (ObjectPrinter, error) {
	switch format {
	case FormatYAML:
		return &YAMLObjectPrinter{}, nil
	case FormatJSON:
		return JSONObjectPrinter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q, supported formats are yaml, json", format)
	}
}

type YAMLObjectPrinter struct {
	printCount int
}

func (y *YAMLObjectPrinter) PrintObject(obj runtime.Object, w io.Writer) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return err
	}

	y.printCount++
	if y.printCount > 1 {
		_, err := w.Write([]byte("---\n"))
		if err != nil {
			return err
		}
	}

	_, err = w.Write(data)
	return err
}

type JSONObjectPrinter struct{}

func (j JSONObjectPrinter) PrintObject(obj runtime.Object, w io.Writer) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, data, "", "    ")
	if err != nil {
		return err
	}
	buf.WriteRune('\n')

	_, err = w.Write(buf.Bytes())
	return err
}
