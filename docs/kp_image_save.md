## kp image save

Create or patch an image configuration

### Synopsis

Create or patch an image configuration by providing command line arguments.
This image will be created only if it does not exist in the provided namespace, otherwise it will be patched.

The --tag flag is required for a create but is immutable and will be ignored for a patch.
The --cache-size flag can only be used to create or increase the size of the existing cache.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry provided for the image tag.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 ...".

```
kp image save <name> --tag <tag> [flags]
```

### Examples

```
kp image create my-image --tag my-registry.com/my-repo --git https://my-repo.com/my-app.git --git-revision my-branch
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob
kp image save my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code
kp image save my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code --builder my-builder -n my-namespace
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --env food=apple
```

### Options

```
      --blob string                    source code blob url
  -b, --builder string                 builder name
      --cache-size string              cache size as a kubernetes quantity (default "2G")
  -c, --cluster-builder string         cluster builder name
      --dry-run                        only print the object that would be sent, without sending it
      --env stringArray                build time environment variables
      --git string                     git repository url
      --git-revision string            git revision (default "master")
  -h, --help                           help for save
      --local-path string              path to local source code
  -n, --namespace string               kubernetes namespace
      --output string                  output format. supported formats are: yaml, json
      --registry-ca-cert-path string   add CA certificates for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
      --sub-path string                build code at the sub path located within the source code directory
  -t, --tag string                     registry location where the image will be created
  -w, --wait                           wait for image create to be reconciled and tail resulting build logs
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

