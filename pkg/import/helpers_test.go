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

	when("registry credentials are provided by environment variables and Get is called", func() {
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
			credHelper = importpkg.NewCredHelper(os.Getenv(envVarRegistryUser), os.Getenv(envVarRegistryPassword))
			registryUser, registryPassword, err := credHelper.Get(envVarRegistryUrl)
			require.NoError(t, err)
			require.Equal(t, "some-registry-user", registryUser)
			require.Equal(t, "some-registry-password", registryPassword)
		})
	})
}
