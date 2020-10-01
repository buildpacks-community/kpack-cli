package commands

import (
	"fmt"
	"github.com/aryann/difflib"
	"github.com/ghodss/yaml"
	"github.com/mgutz/ansi"
	"strings"
)

type Diffable interface {
	DiffObject() interface{}
}

type EmptyDiffer struct {
}

func (ed EmptyDiffer) DiffObject() interface{} {
	return ""
}

type Differ struct {
}

func (d Differ) Diff(dOld, dNew Diffable) (diff string, err error) {
	dataOld := []byte("")

	if dOld.DiffObject() != "" {
		dataOld, err = yaml.Marshal(dOld.DiffObject())
		if err != nil {
			return "", err
		}
	}
	dataNew, err := yaml.Marshal(dNew.DiffObject())
	if err != nil {
		return "", err
	}

	if string(dataOld) == string(dataNew) {
		return "", err
	}

	return renderDiff(string(dataOld), string(dataNew)), nil
}

func renderDiff(a, b string) string {
	diffs := difflib.Diff(strings.Split(a, "\n"), strings.Split(b, "\n"))

	sBuilder := strings.Builder{}
	for _, diff := range diffs {
		text := diff.Payload
		if text == "" {
			continue
		}

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

//func DiffPrettyText(diffs []diffmatchpatch.Diff) string {
//	var buff bytes.Buffer
//	for _, diff := range diffs {
//		text := diff.Text
//
//		switch diff.Type {
//		case diffmatchpatch.DiffInsert:
//			_, _ = buff.WriteString("+ \x1b[32m")
//			_, _ = buff.WriteString(text)
//			_, _ = buff.WriteString("\x1b[0m")
//		case diffmatchpatch.DiffDelete:
//			_, _ = buff.WriteString("- \x1b[31m")
//			_, _ = buff.WriteString(text)
//			_, _ = buff.WriteString("\x1b[0m")
//		case diffmatchpatch.DiffEqual:
//			_, _ = buff.WriteString("  " + text)
//		}
//	}
//
//	return buff.String()
//}
