// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"io"
	"io/ioutil"
	"os"

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
