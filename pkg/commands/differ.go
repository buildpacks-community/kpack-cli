package commands

import (
	"fmt"
	"strings"

	"github.com/aryann/difflib"
	"github.com/ghodss/yaml"
	"github.com/mgutz/ansi"
)

type Differ struct {
}

func (d Differ) Diff(dOld, dNew interface{}) (string, error) {
	dataOld, err := getData(dOld)
	if err != nil {
		return "", err
	}

	dataNew, err := getData(dNew)
	if err != nil {
		return "", err
	}

	if dataOld == dataNew {
		return "", nil
	}

	return renderDiff(dataOld, dataNew), nil
}

func getData(obj interface{}) (string, error) {
	if obj == nil {
		return "", nil
	}
	if str, ok := obj.(string); ok {
		return str, nil
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
