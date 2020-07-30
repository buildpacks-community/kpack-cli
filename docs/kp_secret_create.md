## kp secret create

Create a secret configuration

### Synopsis

Create a secret configuration using registry or git credentials in the provided namespace.

namespace defaults to the kubernetes current-context namespace.

The flags for this command determine the type of secret that will be created:

  "--dockerhub" to create DockerHub credentials.
  Use the "DOCKER_PASSWORD" env var to bypass the password prompt.

  "--gcr" to create Google Container Registry credentials.
  Alternatively, provided the credentials in the "GCR_SERVICE_ACCOUNT_PATH" env var instead of the "--gcr" flag.

  "--registry" and "--registry-user" to create credentials for other registries.
  Use the "REGISTRY_PASSWORD" env var to bypass the password prompt.

  "--git" and "--git-ssh-key" to create SSH based git credentials.
  Alternatively, provided the credentials in the "GIT_SSH_KEY_PATH" env var instead of the "--git-ssh-key" flag.

  "--git" and "--git-user" to create Basic Auth based git credentials.
  Use the "GIT_PASSWORD" env var to bypass the password prompt.

```
kp secret create <name> [flags]
```

### Examples

```
kp secret create my-docker-hub-creds --dockerhub dockerhub-id
kp secret create my-gcr-creds --gcr /path/to/gcr/service-account.json
kp secret create my-registry-cred --registry example-registry.io/my-repo --registry-user my-registry-user
kp secret create my-git-ssh-cred --git git@github.com --git-ssh-key /path/to/git/ssh-private-key.pem
kp secret create my-git-cred --git https://github.com --git-user my-git-user
```

### Options

```
      --dockerhub string       dockerhub id
      --gcr string             path to a file containing the GCR service account
      --git string             git url
      --git-ssh-key string     path to a file containing the Git SSH private key
      --git-user string        git user
  -h, --help                   help for create
  -n, --namespace string       kubernetes namespace
      --registry string        registry
      --registry-user string   registry user
```

### SEE ALSO

* [kp secret](kp_secret.md)	 - Secret Commands

