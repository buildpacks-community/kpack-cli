package dockercreds_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/dockercreds"
)

func TestCredHelper(t *testing.T) {
	spec.Run(t, "TestCredHelper", testCredHelper)
}

const envVarRegistryUrl = "REGISTRY_URL_TEST_VAR"
const envVarRegistryUser = "REGISTRY_USER_TEST_VAR"
const envVarRegistryPassword = "REGISTRY_PASSWORD_TEST_VAR"

type registryVariable struct {
	URLName       string
	URLValue      string
	UserName      string
	UserValue     string
	PasswordName  string
	PasswordValue string
}

func getValuesUnderTest() []registryVariable {
	values := make([]registryVariable, 0, 1+len(digitToWord()))
	values = append(values, registryVariable{
		URLName:       envVarRegistryUrl,
		URLValue:      "some-registry.io",
		UserName:      envVarRegistryUser,
		UserValue:     "some-registry-user",
		PasswordName:  envVarRegistryPassword,
		PasswordValue: "some-registry-password",
	})
	for digit, word := range digitToWord() {
		values = append(values, registryVariable{
			URLName:       envVarRegistryUrl + "_" + strconv.Itoa(digit),
			URLValue:      word + "-registry.io",
			UserName:      envVarRegistryUser + "_" + strconv.Itoa(digit),
			UserValue:     word + "-registry-user",
			PasswordName:  envVarRegistryPassword + "_" + strconv.Itoa(digit),
			PasswordValue: word + "-registry-password",
		})
	}
	return values
}

func digitToWord() map[int]string {
	return map[int]string{
		0:   "zero",
		1:   "one",
		2:   "two",
		3:   "three",
		9:   "nine",
		10:  "ten",
		111: "one-hundred-eleven",
	}
}

func setRegistryEnvVars(t *testing.T) {
	for _, val := range getValuesUnderTest() {
		require.NoError(t, os.Setenv(val.URLName, val.URLValue))
		require.NoError(t, os.Setenv(val.UserName, val.UserValue))
		require.NoError(t, os.Setenv(val.PasswordName, val.PasswordValue))
	}
}

func unsetRegistryEnvVars(t *testing.T) {
	for _, val := range getValuesUnderTest() {
		require.NoError(t, os.Unsetenv(val.URLName))
		require.NoError(t, os.Unsetenv(val.UserName))
		require.NoError(t, os.Unsetenv(val.PasswordName))
	}
}

func testCredHelper(t *testing.T, when spec.G, it spec.S) {

	when("one registry credential is provided by environment variables and Get is called", func() {
		var credHelper *dockercreds.CredHelper

		it.Before(func() {
			require.NoError(t, os.Setenv(envVarRegistryUrl, "foo-registry.io"))
			require.NoError(t, os.Setenv(envVarRegistryUser, "foo-registry-user"))
			require.NoError(t, os.Setenv(envVarRegistryPassword, "foo-registry-password"))
		})

		it.After(func() {
			require.NoError(t, os.Unsetenv(envVarRegistryUrl))
			require.NoError(t, os.Unsetenv(envVarRegistryUser))
			require.NoError(t, os.Unsetenv(envVarRegistryPassword))
		})

		it("returns username and password provided by environment variables", func() {
			credHelper = dockercreds.NewCredHelperFromEnvVars(envVarRegistryUrl, envVarRegistryUser, envVarRegistryPassword)
			require.Equal(t, len(credHelper.Auths), 1)
			registryUser, registryPassword, err := credHelper.Get(os.Getenv(envVarRegistryUrl))
			require.NoError(t, err)
			require.Equal(t, "foo-registry-user", registryUser)
			require.Equal(t, "foo-registry-password", registryPassword)
		})
	})

	when("many registry credentials are provided by environment variables and Get is called", func() {
		var credHelper *dockercreds.CredHelper

		it.Before(func() {
			setRegistryEnvVars(t)
		})

		it.After(func() {
			unsetRegistryEnvVars(t)
		})

		it("returns username and password provided by environment variables", func() {
			credHelper = dockercreds.NewCredHelperFromEnvVars(envVarRegistryUrl, envVarRegistryUser, envVarRegistryPassword)
			require.Equal(t, len(credHelper.Auths), len(getValuesUnderTest()))
			registryUser, registryPassword, err := credHelper.Get(os.Getenv(envVarRegistryUrl + "_3"))
			require.NoError(t, err)
			require.Equal(t, "three-registry-user", registryUser)
			require.Equal(t, "three-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl + "_111"))
			require.NoError(t, err)
			require.Equal(t, "one-hundred-eleven-registry-user", registryUser)
			require.Equal(t, "one-hundred-eleven-registry-password", registryPassword)
			registryUser, registryPassword, err = credHelper.Get(os.Getenv(envVarRegistryUrl))
			require.NoError(t, err)
			require.Equal(t, "some-registry-user", registryUser)
			require.Equal(t, "some-registry-password", registryPassword)
		})
	})
}
