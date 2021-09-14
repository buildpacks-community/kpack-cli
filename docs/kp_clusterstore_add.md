## kp clusterstore add

Add buildpackage(s) to cluster store

### Synopsis

Upload buildpackage(s) to a specific cluster-scoped buildpack store.

Buildpackages will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.


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
      --dry-run                        perform validation with no side-effects; no objects are sent to the server.
                                         The --dry-run flag can be used in combination with the --output flag to
                                         view the Kubernetes resource(s) without sending anything to the server.
      --dry-run-with-image-upload      similar to --dry-run, but with container image uploads allowed.
                                         This flag is provided as a convenience for kp commands that can output Kubernetes
                                         resource with generated container image references. A "kubectl apply -f" of the
                                         resource from --output without image uploads will result in a reconcile failure.
  -h, --help                           help for add
      --output string                  print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                         The output can be used with the "kubectl apply -f" command. To allow this, the command 
                                         updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
      --registry-ca-cert-path string   add CA certificate for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
```

### SEE ALSO

* [kp clusterstore](kp_clusterstore.md)	 - ClusterStore Commands

