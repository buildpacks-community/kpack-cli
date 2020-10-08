## kp clusterstore add

Add buildpackage(s) to cluster store

### Synopsis

Upload buildpackage(s) to a specific cluster-scoped buildpack store.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.


```
kp clusterstore add <store> -b <buildpackage> [-b <buildpackage>...] [flags]
```

### Examples

```
kp clusterstore add my-store -b my-registry.com/my-buildpackage
kp clusterstore add my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage -b my-registry.com/my-third-buildpackage
kp clusterstore add my-store -b ../path/to/my-local-buildpackage.cnb
```

### Options

```
  -b, --buildpackage stringArray       location of the buildpackage
      --dry-run                        only print the object that would be sent, without sending it
  -h, --help                           help for add
      --output string                  output format. supported formats are: yaml, json
      --registry-ca-cert-path string   add CA certificates for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
```

### SEE ALSO

* [kp clusterstore](kp_clusterstore.md)	 - ClusterStore Commands

