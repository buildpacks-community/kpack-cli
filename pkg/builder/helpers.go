// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ghodss/yaml"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

func ReadOrder(path string) ([]v1alpha1.OrderEntry, error) {
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

	var order []v1alpha1.OrderEntry
	return order, yaml.Unmarshal(buf, &order)
}

func CreateOrder(buildpacks []string) []v1alpha1.OrderEntry {
	group := make([]v1alpha1.BuildpackRef, 0)

	// this regular expression splits out buildpack id and version
	var re = regexp.MustCompile(`(?m)^([^@]+)[@]?(.*)`)

	for _, buildpack := range buildpacks {
		submatch := re.FindStringSubmatch(buildpack)

		id := submatch[1]
		version := submatch[2]

		group = append(group, v1alpha1.BuildpackRef{
			BuildpackInfo: v1alpha1.BuildpackInfo{
				Id:      id,
				Version: version,
			},
		})
	}

	return []v1alpha1.OrderEntry{{Group: group}}
}
