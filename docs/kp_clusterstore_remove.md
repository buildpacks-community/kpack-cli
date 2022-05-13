## kp clusterstore remove

Remove buildpackage(s) from cluster store

### Synopsis

Removes existing buildpackage(s) from a specific cluster-scoped buildpack store.


```
kp clusterstore remove <store> -b <buildpackage> [-b <buildpackage>...] [flags]
```

### Examples

```
kp clusterstore remove my-store -b buildpackage@1.0.0
kp clusterstore remove my-store -b buildpackage@1.0.0 -b other-buildpackage@2.0.0

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
                                     The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
```

### SEE ALSO

* [kp clusterstore](kp_clusterstore.md)	 - ClusterStore Commands

