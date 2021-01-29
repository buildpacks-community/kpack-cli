## kp lifecycle update

Update lifecycle image used by kpack

### Synopsis

Update lifecycle image used by kpack

The Lifecycle image will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

The canonical repository is read from the "canonical.repository" key of the "kp-config" ConfigMap within "kpack" namespace.


```
kp lifecycle update --image <image-tag> [flags]
```

### Examples

```
kp lifecycle update --image my-registry.com/lifecycle
```

### Options

```
  -h, --help                           help for update
  -i, --image string                   location of the image
      --registry-ca-cert-path string   add CA certificate for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
```

### SEE ALSO

* [kp lifecycle](kp_lifecycle.md)	 - Lifecycle Commands

