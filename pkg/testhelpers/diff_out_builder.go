// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mgutz/ansi"
	"github.com/stretchr/testify/assert"
)

const defaultPrefix = ""

type DiffBuilder struct {
	t  *testing.T
	sb strings.Builder
	o  DiffOptions
}

type DiffOptions struct {
	Prefix         string
	Color          bool
	TrimOutNewline bool
}

func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		Prefix:         defaultPrefix,
		Color:          true,
		TrimOutNewline: true,
	}
}

func NewDiffBuilder(t *testing.T) *DiffBuilder {
	builder := &DiffBuilder{t: t, sb: strings.Builder{}}
	builder.Configure(DefaultDiffOptions())
	return builder
}

func (d *DiffBuilder) Configure(options DiffOptions) *DiffBuilder {
	d.o = options
	return d
}

func (d *DiffBuilder) SetPrefix(prefix string) *DiffBuilder {
	d.o.Prefix = prefix
	return d
}

func (d *DiffBuilder) Reset() *DiffBuilder {
	d.sb.Reset()
	return d
}

func (d *DiffBuilder) Txt(str string) *DiffBuilder {
	textLine := fmt.Sprintf("%s\n", str)
	_, err := d.sb.WriteString(textLine)
	assert.NoError(d.t, err)
	return d
}

func (d *DiffBuilder) NoD(str string) *DiffBuilder {
	noDiffLine := fmt.Sprintf("%s%s\n", d.o.Prefix, str)
	_, err := d.sb.WriteString(noDiffLine)
	assert.NoError(d.t, err)
	return d
}

func (d *DiffBuilder) Old(str string) *DiffBuilder {
	var oldLine string
	if d.o.Color {
		oldLine = fmt.Sprintf("%s%s %s\n", d.o.Prefix, ansi.Color("-", "red"), ansi.Color(str, "red"))
	} else {
		oldLine = fmt.Sprintf("%s%s %s\n", d.o.Prefix, "-", str)
	}
	_, err := d.sb.WriteString(oldLine)
	assert.NoError(d.t, err)
	return d
}

func (d *DiffBuilder) New(str string) *DiffBuilder {
	var newLine string
	if d.o.Color {
		newLine = fmt.Sprintf("%s%s %s\n", d.o.Prefix, ansi.Color("+", "green"), ansi.Color(str, "green"))
	} else {
		newLine = fmt.Sprintf("%s%s %s\n", d.o.Prefix, "+", str)
	}
	_, err := d.sb.WriteString(newLine)
	assert.NoError(d.t, err)
	return d
}

func (d *DiffBuilder) Out() string {
	outStr := d.sb.String()
	if d.o.TrimOutNewline {
		outStr = strings.TrimSuffix(outStr, "\n")
	}
	return outStr
}
