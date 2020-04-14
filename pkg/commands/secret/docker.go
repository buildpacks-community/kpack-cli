package secret

import "github.com/google/go-containerregistry/pkg/authn"

type DockerCreds map[string]authn.AuthConfig

type dockerConfigJson struct {
	Auths DockerCreds `json:"auths"`
}
