## Authentication Via Environment Variables

`kp` can also use following environment variables for registry authentication:

- `KP_REGISTRY_HOSTNAME` to specify registry URL (e.g. gcr.io, docker.io, my-harbor-instance.net)
- `KP_REGISTRY_USERNAME` to specify registry username
- `KP_REGISTRY_PASSWORD` to specify registry password

Since you may need to provide multiple registry credentials, you may use the above environment variables multiple times with a suffix of `_N` where N is a positive integer. Use same suffix for registry URL, username and password.

#### Example

```bash
$ KP_REGISTRY_HOSTNAME_1=gcr.io \
    KP_REGISTRY_USERNAME_1=pat \
    KP_REGISTRY_PASSWORD_1=p4ssw0rd \
    KP_REGISTRY_HOSTNAME_2=docker.io \
    KP_REGISTRY_USERNAME_2=sam \
    KP_REGISTRY_PASSWORD_2=s3cret \
    kp import -f descriptor.yaml
```

Credentials provided by these environment variables will be used first when authenticating against a given registry. If those credentials fail, the credentials provided in your `~/.docker/config.json` will be used.

### Affected `kp` commands

The following `kp` commands can utilize environment variables for registry authentication:

  * `kp clusterstack create`
  * `kp clusterstack patch`
  * `kp clusterstack save`
  * `kp clusterstore add`
  * `kp clusterstore create`
  * `kp clusterstore save`
  * `kp lifecycle patch`
  * `kp image create`
  * `kp image patch`
  * `kp image save`
  * `kp import`