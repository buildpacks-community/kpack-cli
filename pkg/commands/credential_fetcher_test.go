// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
)

func TestCredentialFetcher(t *testing.T) {
	spec.Run(t, "TestCredentialFetcher", testCredentialFetcher)
}

func testCredentialFetcher(t *testing.T, when spec.G, it spec.S) {
	when("an environment variable is provided", func() {
		const envVar = "SOME_TEST_ENV_VAR"

		it.Before(func() {
			require.NoError(t, os.Setenv(envVar, "some-password-value"))
		})

		it.After(func() {
			_ = os.Setenv(envVar, "")
		})

		it("reads the password from the env var", func() {
			password, err := commands.CredentialFetcher{}.FetchPassword(envVar, "")
			require.NoError(t, err)
			require.Equal(t, "some-password-value", password)
		})
	})
}
