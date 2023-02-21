package dockercreds_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/dockercreds"
)

func TestKeychainHelper(t *testing.T) {
	spec.Run(t, "TestKeychainHelper", testKeychainHelper)
}

type FakeResource struct {
	string      string
	registryStr string
}

func (r FakeResource) String() string {
	return r.string
}

func (r FakeResource) RegistryStr() string {
	return r.registryStr
}

func encode(user, pass string) string {
	delimited := fmt.Sprintf("%s:%s", user, pass)
	return base64.StdEncoding.EncodeToString([]byte(delimited))
}

func testKeychainHelper(t *testing.T, when spec.G, it spec.S) {

	when("many username and password are provided by default environment variables and Resolve is called", func() {
		it.Before(func() {
			require.NoError(t, os.Setenv(dockercreds.EnvVarRegistryUrl, "foo-registry.io"))
			require.NoError(t, os.Setenv(dockercreds.EnvVarRegistryUser, "foo-registry-user"))
			require.NoError(t, os.Setenv(dockercreds.EnvVarRegistryPassword, "foo-registry-password"))
		})

		it.After(func() {
			require.NoError(t, os.Unsetenv(dockercreds.EnvVarRegistryUrl))
			require.NoError(t, os.Unsetenv(dockercreds.EnvVarRegistryUser))
			require.NoError(t, os.Unsetenv(dockercreds.EnvVarRegistryPassword))
		})

		it("returns AuthConfig with provided username and password", func() {
			content := fmt.Sprintf(`{"auths": {"foo-registry.io": {"auth": %q}}}`, encode("foo-registry-user", "foo-registry-password"))
			resource := &FakeResource{
				string:      content,
				registryStr: "foo-registry.io",
			}

			keychain := dockercreds.NewKeychainFromDefaultEnvVarsWithDefault().Keychain

			expected := &authn.AuthConfig{
				Username: "foo-registry-user",
				Password: "foo-registry-password",
			}

			auth, _ := keychain.Resolve(resource)
			cfg, _ := auth.Authorization()

			require.Equal(t, cfg.Username, expected.Username)
			require.Equal(t, cfg.Password, expected.Password)
		})
	})

}
