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
	FormatYAML    string = "yaml"
	FormatJSON    string = "json"
)

type ResourcePrinter interface {
	PrintObject(obj runtime.Object, w io.Writer) error
}

func NewObjectPrinter(format string) (ResourcePrinter, error) {
	switch format {
	case FormatYAML:
		return &YAMLResourcePrinter{}, nil
	case FormatJSON:
		return JSONResourcePrinter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s, supported formats are yaml, json", format)
	}
}

type YAMLResourcePrinter struct {
	printCount int
}

func (y *YAMLResourcePrinter) PrintObject(obj runtime.Object, w io.Writer) error {
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

type JSONResourcePrinter struct {}

func (j JSONResourcePrinter) PrintObject(obj runtime.Object, w io.Writer) error {
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
