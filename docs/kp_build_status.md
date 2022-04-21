## kp build status

Display status for an image resource build

### Synopsis

Prints detailed information about the status of a specific build of an image resource in the provided namespace.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.

When using the --bom flag, only the built image's bill of materials will be printed.
Using the --bom flag will read metadata from the build's built image in the registry
Therefore, you must have credentials to access the registry on your machine when using the --bom flag.
--registry-ca-cert-path and --registry-verify-certs are only used when using the --bom flag.

```
kp build status <image-name> [flags]
```

### Examples

```
kp build status my-image
kp build status my-image -b 2 -n my-namespace
```

### Options

```
      --bom                            only print the built image bill of materials
  -b, --build string                   build number
  -h, --help                           help for status
  -n, --namespace string               kubernetes namespace
      --registry-ca-cert-path string   add CA certificate for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

