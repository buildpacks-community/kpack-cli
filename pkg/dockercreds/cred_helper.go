package dockercreds

import (
	"fmt"
	"os"
	"regexp"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
)

type CredHelper struct {
	Auths map[string]authn.Basic
}

const (
	EnvVarRegistryUrl      = "KP_REGISTRY_HOSTNAME"
	EnvVarRegistryUser     = "KP_REGISTRY_USERNAME"
	EnvVarRegistryPassword = "KP_REGISTRY_PASSWORD"
)

var (
	DefaultKeychain authn.Keychain = authn.NewMultiKeychain(authn.NewKeychainFromHelper(
		newCredHelperFromEnvVars(EnvVarRegistryUrl, EnvVarRegistryUser, EnvVarRegistryPassword)),
		authn.DefaultKeychain)
)

func newCredHelperFromEnvVars(serverURLEnvVar string, usernameEnvVar string, passwordEnvVar string) *CredHelper {
	auths := map[string]authn.Basic{}

	var registryRegex = regexp.MustCompile(fmt.Sprintf(`(%s(_\d+)?)=(.*)`, serverURLEnvVar))
	for _, env := range os.Environ() {
		if match := registryRegex.FindStringSubmatch(env); len(match) > 0 {
			auths[match[3]] = authn.Basic{
				Username: os.Getenv(usernameEnvVar + match[2]),
				Password: os.Getenv(passwordEnvVar + match[2]),
			}
		}
	}

	return &CredHelper{
		Auths: auths,
	}
}

func (c *CredHelper) Get(serverURL string) (string, string, error) {
	auth, found := c.Auths[serverURL]
	if found == false {
		return "", "", errors.New("serverURL does not refer to a known registry")
	} else {
		return auth.Username, auth.Password, nil
	}
}
