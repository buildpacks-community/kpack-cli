package commands_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func TestPasswordReader(t *testing.T) {
	spec.Run(t, "TestPasswordReader", testPasswordReader)
}

func testPasswordReader(t *testing.T, when spec.G, it spec.S) {
	when("an environment variable is provied", func() {

		const envVar = "SOME_TEST_ENV_VAR"

		it.Before(func() {
			require.NoError(t, os.Setenv(envVar, "some-password-value"))
		})

		it.After(func() {
			_ = os.Setenv(envVar, "")
		})

		it("reads the password from the env var", func() {
			password, err := commands.PasswordReader{}.Read(nil, "", envVar)
			require.NoError(t, err)
			require.Equal(t, "some-password-value", password)
		})
	})
}
