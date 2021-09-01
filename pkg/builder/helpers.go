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
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

func ReadOrder(path string) ([]corev1alpha1.OrderEntry, error) {
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

	var order []corev1alpha1.OrderEntry
	return order, yaml.Unmarshal(buf, &order)
}

func CreateOrder(buildpacks []string) []corev1alpha1.OrderEntry {
	group := make([]corev1alpha1.BuildpackRef, 0)

	// this regular expression splits out buildpack id and version
	var re = regexp.MustCompile(`(?m)^([^@]+)[@]?(.*)`)

	for _, buildpack := range buildpacks {
		submatch := re.FindStringSubmatch(buildpack)

		id := submatch[1]
		version := submatch[2]

		group = append(group, corev1alpha1.BuildpackRef{
			BuildpackInfo: corev1alpha1.BuildpackInfo{
				Id:      id,
				Version: version,
			},
		})
	}

	return []corev1alpha1.OrderEntry{{Group: group}}
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
