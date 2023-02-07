package _import

import (
	"os"
	"strconv"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
)

type CredHelper struct {
	Auths map[string]authn.Basic
}

func NewCredHelperFromEnvVars(serverURLEnvVar string, usernameEnvVar string, passwordEnvVar string) *CredHelper {
	suffixes := [11]string{""}
	for i := 0; i <= 9; i++ {
		suffixes[i+1] = "_" + strconv.Itoa(i)
	}

	auths := map[string]authn.Basic{}

	for _, suffix := range suffixes {
		if serverUrl := os.Getenv(serverURLEnvVar + suffix); serverUrl != "" {
			auths[serverUrl] = authn.Basic{
				Username: os.Getenv(usernameEnvVar + suffix),
				Password: os.Getenv(passwordEnvVar + suffix),
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
