// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
)

func CreatePatch(original, updated interface{}) ([]byte, error) {
	originalBytes, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}

	updatedBytes, err := json.Marshal(updated)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreateMergePatch(originalBytes, updatedBytes)
	if err != nil {
		return nil, err
	}

	if string(patch) == "{}" {
		return nil, nil
	}

	return patch, nil
}
