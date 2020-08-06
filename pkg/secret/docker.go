// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret

import "github.com/google/go-containerregistry/pkg/authn"

type DockerCredentials map[string]authn.AuthConfig

type DockerConfigJson struct {
	Auths DockerCredentials `json:"auths"`
}
