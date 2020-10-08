## kp image patch

Patch an existing image configuration

### Synopsis

Patch an existing image configuration by providing command line arguments.
This will fail if the image does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry as the existing image tag.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".

Existing environment variables may be deleted by using the "--delete-env" flag.
For each environment variable, supply the "--delete-env" flag followed by the variable name.
For example, "--delete-env key1 --delete-env key2 ...".

The --cache-size flag can only be used to increase the size of the existing cache.


```
kp image patch <name> [flags]
```

### Examples

```
kp image patch my-image --git-revision my-other-branch
kp image patch my-image --blob https://my-blob-host.com/my-blob
kp image patch my-image --local-path /path/to/local/source/code
kp image patch my-image --local-path /path/to/local/source/code --builder my-builder
kp image patch my-image --env foo=bar --env color=red --delete-env apple --delete-env potato
```

### Options

```
      --blob string                    source code blob url
      --builder string                 builder name
      --cache-size string              cache size as a kubernetes quantity
      --cluster-builder string         cluster builder name
  -d, --delete-env stringArray         build time environment variables to remove
      --dry-run                        only print the object that would be sent, without sending it
  -e, --env stringArray                build time environment variables to add/replace
      --git string                     git repository url
      --git-revision string            git revision (default "master")
  -h, --help                           help for patch
      --local-path string              path to local source code
  -n, --namespace string               kubernetes namespace
      --output string                  output format. supported formats are: yaml, json
      --registry-ca-cert-path string   add CA certificates for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
      --sub-path string                build code at the sub path located within the source code directory
  -w, --wait                           wait for image patch to be reconciled and tail resulting build logs
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

