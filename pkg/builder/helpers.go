// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ghodss/yaml"
	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

func ReadOrder(path string) ([]buildv1alpha2.BuilderOrderEntry, error) {
	var (
		file io.ReadCloser
		err  error
	)

	if path == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var order []buildv1alpha2.BuilderOrderEntry
	return order, yaml.Unmarshal(buf, &order)
}

func CreateOrder(buildpacks []string) []buildv1alpha2.BuilderOrderEntry {
	group := make([]buildv1alpha2.BuilderBuildpackRef, 0)

	// this regular expression splits out buildpack id and version
	var re = regexp.MustCompile(`(?m)^([^@]+)[@]?(.*)`)

	for _, buildpack := range buildpacks {
		submatch := re.FindStringSubmatch(buildpack)

		id := submatch[1]
		version := submatch[2]

		group = append(group, buildv1alpha2.BuilderBuildpackRef{
			BuildpackRef: corev1alpha1.BuildpackRef{
				BuildpackInfo: corev1alpha1.BuildpackInfo{
					Id:      id,
					Version: version,
				},
			},
		})
	}

	return []buildv1alpha2.BuilderOrderEntry{{Group: group}}
}

func CreateDetectionOrderRow(ref corev1alpha1.BuildpackRef) (string, string) {
	data := fmt.Sprintf("  %s", ref.Id)
	optional := ""

	if ref.Version != "" {
		data = fmt.Sprintf("%s@%s", data, ref.Version)
	}

	if ref.Optional {
		optional = "(Optional)"
	}

	return data, optional
}

func CoreOrderEntryToBuildOrderEntry(order []corev1alpha1.OrderEntry) []buildv1alpha2.BuilderOrderEntry {
	res := make([]buildv1alpha2.BuilderOrderEntry, len(order))
	for i, entry := range order {
		group := make([]buildv1alpha2.BuilderBuildpackRef, len(entry.Group))
		for j, ref := range entry.Group {
			group[j] = buildv1alpha2.BuilderBuildpackRef{
				BuildpackRef: ref,
			}
		}

		res[i] = buildv1alpha2.BuilderOrderEntry{
			Group: group,
		}
	}
	return res
}
