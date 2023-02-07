package _import_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	importpkg "github.com/vmware-tanzu/kpack-cli/pkg/import"
)

func TestCredHelper(t *testing.T) {
	spec.Run(t, "TestCredHelper", testCredHelper)
}

func testCredHelper(t *testing.T, when spec.G, it spec.S) {
	const envVarRegistryUrl = "REGISTRY_URL_TEST_VAR"
	const envVarRegistryUser = "REGISTRY_USER_TEST_VAR"
	const envVarRegistryPassword = "REGISTRY_PASSWORD_TEST_VAR"

	when("one registry credential is provided by environment variables and Get is called", func() {
		var credHelper *importpkg.CredHelper

		it.Before(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl, "some-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser, "some-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword, "some-registry-password"))
		})

		it.After(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl, ""))
			require.NoError(t, os.Setenv(envVarRegistryUser, ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword, ""))
		})

		it("returns username and password provided by environment variables", func() {
			credHelper = importpkg.NewCredHelperFromEnvVars(envVarRegistryUrl, envVarRegistryUser, envVarRegistryPassword)
			require.Equal(t, len(credHelper.Auths), 1)
			registryUser, registryPassword, err := credHelper.Get(os.Getenv(envVarRegistryUrl))
			require.NoError(t, err)
			require.Equal(t, "some-registry-user", registryUser)
			require.Equal(t, "some-registry-password", registryPassword)
		})
	})

	when("many registry credentials are provided by environment variables and Get is called", func() {
		var credHelper *importpkg.CredHelper

		it.Before(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl+"_0", "fizz-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_0", "fizz-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_0", "fizz-registry-password"))
			require.NoError(t, os.Setenv(envVarRegistryUrl+"_1", "fuzz-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_1", "fuzz-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_1", "fuzz-registry-password"))
			require.NoError(t, os.Setenv(envVarRegistryUrl, "some-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser, "some-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword, "some-registry-password"))
		})

		it.After(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl+"_0", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_0", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_0", ""))
			require.NoError(t, os.Setenv(envVarRegistryUrl+"_1", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_1", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_1", ""))
			require.NoError(t, os.Setenv(envVarRegistryUrl, ""))
			require.NoError(t, os.Setenv(envVarRegistryUser, ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword, ""))
		})

		it("returns username and password provided by environment variables", func() {
			credHelper = importpkg.NewCredHelperFromEnvVars(envVarRegistryUrl, envVarRegistryUser, envVarRegistryPassword)
			require.Equal(t, len(credHelper.Auths), 3)
			registryUser, registryPassword, err := credHelper.Get(os.Getenv(envVarRegistryUrl + "_0"))
			require.NoError(t, err)
			require.Equal(t, "fizz-registry-user", registryUser)
			require.Equal(t, "fizz-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl + "_1"))
			require.NoError(t, err)
			require.Equal(t, "fuzz-registry-user", registryUser)
			require.Equal(t, "fuzz-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl))
			require.NoError(t, err)
			require.Equal(t, "some-registry-user", registryUser)
			require.Equal(t, "some-registry-password", registryPassword)
		})
	})

	when("more than 10 registry credentials are provided by environment variables that "+
		"have a \"_N\" suffix (e.g., REGISTRY_URL_N where N is an integer >= 0 and <= 9) "+
		"and no credential is provided by environment variables with an empty suffix (e.g., REGISTRY_URL) "+
		"and Get is called", func() {
		var credHelper *importpkg.CredHelper

		it.Before(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl+"_0", "zero-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_0", "zero-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_0", "zero-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_1", "one-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_1", "one-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_1", "one-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_2", "two-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_2", "two-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_2", "two-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_3", "three-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_3", "three-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_3", "three-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_4", "four-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_4", "four-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_4", "four-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_5", "five-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_5", "five-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_5", "five-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_6", "six-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_6", "six-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_6", "six-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_7", "seven-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_7", "seven-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_7", "seven-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_8", "eight-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_8", "eight-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_8", "eight-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_9", "nine-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_9", "nine-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_9", "nine-registry-password"))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_10", "ten-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_10", "ten-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_10", "ten-registry-password"))
		})

		it.After(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl+"_0", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_0", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_0", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_1", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_1", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_1", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_2", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_2", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_2", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_3", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_3", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_3", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_4", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_4", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_4", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_5", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_5", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_5", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_6", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_6", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_6", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_7", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_7", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_7", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_8", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_8", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_8", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_9", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_9", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_9", ""))

			require.NoError(t, os.Setenv(envVarRegistryUrl+"_10", ""))
			require.NoError(t, os.Setenv(envVarRegistryUser+"_10", ""))
			require.NoError(t, os.Setenv(envVarRegistryPassword+"_10", ""))
		})

		it("returns credentials for REGISTRY_URL_N where N is an integer >= 0 and <= 9", func() {
			credHelper = importpkg.NewCredHelperFromEnvVars(envVarRegistryUrl, envVarRegistryUser, envVarRegistryPassword)
			require.Equal(t, len(credHelper.Auths), 10)
			registryUser, registryPassword, err := credHelper.Get(os.Getenv(envVarRegistryUrl + "_0"))
			require.NoError(t, err)
			require.Equal(t, "zero-registry-user", registryUser)
			require.Equal(t, "zero-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl + "_4"))
			require.NoError(t, err)
			require.Equal(t, "four-registry-user", registryUser)
			require.Equal(t, "four-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl + "_9"))
			require.NoError(t, err)
			require.Equal(t, "nine-registry-user", registryUser)
			require.Equal(t, "nine-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl + "_10"))
			require.ErrorContains(t, err, "serverURL does not refer to a known registry")
		})
	})
}
