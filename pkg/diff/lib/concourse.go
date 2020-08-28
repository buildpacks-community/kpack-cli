package lib

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/aryann/difflib"
	"github.com/mgutz/ansi"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/yaml"
)

type Index interface {
	FindEquivalent(interface{}) (interface{}, bool)
	Slice() []interface{}
}

type Diffs []Diff

type Diff struct {
	Before interface{}
	After  interface{}
}

func name(v interface{}) string {
	return reflect.ValueOf(v).FieldByName("Name").String()
}

func (diff Diff) Render(to io.Writer, label string, changed bool) {

	if diff.Before != nil && diff.After != nil {
		if changed {
			fmt.Fprintf(to, ansi.Color("%s:", "cyan")+"\n", label)
		} else {
			fmt.Fprintf(to, ansi.Color("%s:", "cyan")+"\n", label)

		}

		payloadA, _ := yaml.Marshal(diff.Before)
		payloadB, _ := yaml.Marshal(diff.After)

		renderDiff(to, string(payloadA), string(payloadB))
	} else if diff.Before != nil {
		fmt.Fprintf(to, ansi.Color("%s %s has been removed:", "yellow")+"\n", label, name(diff.Before))

		payloadA, _ := yaml.Marshal(diff.Before)

		renderDiff(to, string(payloadA), "")
	} else {
		fmt.Fprintf(to, ansi.Color("%s %s has been added:", "yellow")+"\n", label, name(diff.After))

		payloadB, _ := yaml.Marshal(diff.After)

		renderDiff(to, "", string(payloadB))
	}
}

func diffIndices(oldIndex Index, newIndex Index) Diffs {
	diffs := Diffs{}

	for _, thing := range oldIndex.Slice() {
		newThing, found := newIndex.FindEquivalent(thing)
		if !found {
			diffs = append(diffs, Diff{
				Before: thing,
				After:  nil,
			})
			continue
		}

		if practicallyDifferent(thing, newThing) {
			diffs = append(diffs, Diff{
				Before: thing,
				After:  newThing,
			})
		}
	}

	for _, thing := range newIndex.Slice() {
		_, found := oldIndex.FindEquivalent(thing)
		if !found {
			diffs = append(diffs, Diff{
				Before: nil,
				After:  thing,
			})
			continue
		}
	}

	return diffs
}

func renderDiff(to io.Writer, a, b string) {
	diffs := difflib.Diff(strings.Split(a, "\n"), strings.Split(b, "\n"))
	indent := gexec.NewPrefixedWriter("\b\b", to)

	for _, diff := range diffs {
		text := diff.Payload

		switch diff.Delta {
		case difflib.RightOnly:
			fmt.Fprintf(indent, "%s %s\n", ansi.Color("+", "green"), ansi.Color(text, "green"))
		case difflib.LeftOnly:
			fmt.Fprintf(indent, "%s %s\n", ansi.Color("-", "red"), ansi.Color(text, "red"))
		case difflib.Common:
			fmt.Fprintf(to, "%s\n", text)
		}
	}
}

func practicallyDifferent(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return false
	}

	// prevent silly things like 300 != 300.0 due to YAML vs. JSON
	// inconsistencies

	marshalledA, _ := yaml.Marshal(a)
	marshalledB, _ := yaml.Marshal(b)

	return !bytes.Equal(marshalledA, marshalledB)
}
