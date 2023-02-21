package dockercreds

import "github.com/google/go-containerregistry/pkg/authn"

const (
	EnvVarRegistryUrl      = "REGISTRY_URL"
	EnvVarRegistryUser     = "REGISTRY_USER"
	EnvVarRegistryPassword = "REGISTRY_PASSWORD"
)

type KeychainHelper struct {
	Keychain authn.Keychain
}

func NewKeychainFromDefaultEnvVarsWithDefault() *KeychainHelper {
	return &KeychainHelper{
		Keychain: authn.NewMultiKeychain(
			authn.NewKeychainFromHelper(
				NewCredHelperFromEnvVars(EnvVarRegistryUrl, EnvVarRegistryUser, EnvVarRegistryPassword)),
			authn.DefaultKeychain),
	}
}
