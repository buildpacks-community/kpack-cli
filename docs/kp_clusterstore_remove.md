## kp clusterstore remove

Remove buildpackage(s) from cluster store

### Synopsis

Removes existing buildpackage(s) from a specific cluster-scoped buildpack store.

This relies on the image(s) specified to exist in the store and removes the associated buildpackage(s)


```
kp clusterstore remove <store> -b <buildpackage> [-b <buildpackage>...] [flags]
```

### Examples

```
kp clusterstore remove my-store -b my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271
kp clusterstore remove my-store -b my-registry.com/my-buildpackage/buildpacks_httpd@sha256:7a09cfeae4763207b9efeacecf914a57e4f5d6c4459226f6133ecaccb5c46271 -b my-registry.com/my-buildpackage/buildpacks_nginx@sha256:eacecf914a57e4f5d6c4459226f6133ecaccb5c462717a09cfeae4763207b9ef

```

### Options

```
  -b, --buildpackage stringArray   buildpackage to remove
      --dry-run                    perform validation with no side-effects; no objects are sent to the server.
                                     The --dry-run flag can be used in combination with the --output flag to
                                     view the Kubernetes resource(s) without sending anything to the server.
  -h, --help                       help for remove
      --output string              print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                     The output can be used with the "kubectl apply -f" command. To allow this, the command 
                                     updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
```

### SEE ALSO

* [kp clusterstore](kp_clusterstore.md)	 - ClusterStore Commands

