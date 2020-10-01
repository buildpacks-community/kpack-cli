package main

import (
	"bytes"
	"fmt"
	"github.com/aryann/difflib"
	"github.com/mgutz/ansi"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strings"
)

const (
	yaml1 = `foo: bar
bar: baz
`
	yaml2 = `foo: bar
bar: newBaz
`
)

func main() {
	//dmp := diffmatchpatch.New()

	//diffs := dmp.DiffMain(yaml1, yaml2, false)
	//fmt.Println(dmp.DiffPrettyText(diffs))

	//fileAdmp, fileBdmp, dmpStrings := dmp.DiffLinesToChars(yaml1, yaml2)
	//diffs := dmp.DiffMain(fileAdmp, fileBdmp, false)
	//diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
	//diffs = dmp.DiffCleanupSemantic(diffs)
	//fmt.Println(DiffPrettyText(diffs))

	//diffs := difflib.Diff(strings.Split(yaml1, "\n"), strings.Split(yaml2, "\n"))
	fmt.Println(renderDiff(yaml1, yaml2))
	fmt.Println(renderDiff("", yaml2))
}

func DiffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("+ \x1b[32m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("- \x1b[31m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("  " + text)
		}
	}

	return buff.String()
}

func renderDiff(a, b string) string {
	diffs := difflib.Diff(strings.Split(a, "\n"), strings.Split(b, "\n"))

	sBuilder := strings.Builder{}
	for _, diff := range diffs {
		text := diff.Payload

		switch diff.Delta {
		case difflib.RightOnly:
			sBuilder.WriteString(fmt.Sprintf("%s %s\n", ansi.Color("+", "green"), ansi.Color(text, "green")))
		case difflib.LeftOnly:
			sBuilder.WriteString(fmt.Sprintf("%s %s\n", ansi.Color("-", "red"), ansi.Color(text, "red")))
		case difflib.Common:
			sBuilder.WriteString(fmt.Sprintf("%s\n", "  "+text))
		}
	}
	return sBuilder.String()
}
