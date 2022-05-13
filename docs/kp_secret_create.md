## kp secret create

Create a secret configuration

### Synopsis

Create a secret configuration using registry or git credentials in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine the type of secret that will be created:

  "--dockerhub" to create DockerHub credentials.
  Use the "DOCKER_PASSWORD" env var to bypass the password prompt.

  "--gcr" to create Google Container Registry credentials.
  Alternatively, provided the credentials in the "GCR_SERVICE_ACCOUNT_PATH" env var instead of the "--gcr" flag.

  "--registry" and "--registry-user" to create credentials for other registries.
  Use the "REGISTRY_PASSWORD" env var to bypass the password prompt.

  "--git-url" and "--git-ssh-key" to create SSH based git credentials.
  "--git-url" should not contain the repository path (eg. git@github.com not git@github.com:my/repo)
  Alternatively, provided the credentials in the "GIT_SSH_KEY_PATH" env var instead of the "--git-ssh-key" flag.

  "--git-url" and "--git-user" to create Basic Auth based git credentials.
  "--git-url" should not contain the repository path (eg. https://github.com not https://github.com/my/repo) 
  Use the "GIT_PASSWORD" env var to bypass the password prompt.

```
kp secret create <name> [flags]
```

### Examples

```
kp secret create my-docker-hub-creds --dockerhub dockerhub-id
kp secret create my-gcr-creds --gcr /path/to/gcr/service-account.json
kp secret create my-registry-cred --registry example-registry.io --registry-user my-registry-user
kp secret create my-git-ssh-cred --git-url git@github.com --git-ssh-key /path/to/git/ssh-private-key.pem
kp secret create my-git-cred --git-url https://github.com --git-user my-git-user
```

### Options

```
      --dockerhub string       dockerhub id
      --dry-run                perform validation with no side-effects; no objects are sent to the server.
                                 The --dry-run flag can be used in combination with the --output flag to
                                 view the Kubernetes resource(s) without sending anything to the server.
      --gcr string             path to a file containing the GCR service account
      --git-ssh-key string     path to a file containing the GitUrl SSH private key
      --git-url string         git url
      --git-user string        git user
  -h, --help                   help for create
  -n, --namespace string       kubernetes namespace
      --output string          print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                 The output can be used with the "kubectl apply -f" command. To allow this, the command
                                 updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
                                 The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
      --registry string        registry
      --registry-user string   registry user
```

### SEE ALSO

* [kp secret](kp_secret.md)	 - Secret Commands

