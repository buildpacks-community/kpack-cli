## kp clusterstack update

Update a cluster stack

### Synopsis

Updates the run and build images of a specific cluster-scoped stack.

The run and build images will be uploaded to the the registry configured on your stack.
Therefore, you must have credentials to access the registry on your machine.

```
kp clusterstack update <name> [flags]
```

### Examples

```
kp clusterstack update my-stack --build-image my-registry.com/build --run-image my-registry.com/run
kp clusterstack update my-stack --build-image ../path/to/build.tar --run-image ../path/to/run.tar
```

### Options

```
  -b, --build-image string             build image tag or local tar file path
      --dry-run                        only print the object that would be sent, without sending it
  -h, --help                           help for update
      --output string                  output format. supported formats are: yaml, json
      --registry-ca-cert-path string   add CA certificates for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
  -r, --run-image string               run image tag or local tar file path
```

### SEE ALSO

* [kp clusterstack](kp_clusterstack.md)	 - ClusterStack Commands

