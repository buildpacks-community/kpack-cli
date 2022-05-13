## kp import

Import dependencies for stores, stacks, and cluster builders

### Synopsis

This operation will create or update clusterstores, clusterstacks, and clusterbuilders defined in the dependency descriptor.

kp import will always attempt to upload the stack, store, and builder images, even if the resources have not changed.
This can be used as a way to repair resources when registry images have been unexpectedly removed.

```
kp import -f <filename> [flags]
```

### Examples

```
kp import -f dependencies.yaml
cat dependencies.yaml | kp import -f -
```

### Options

```
      --dry-run                        perform validation with no side-effects; no objects are sent to the server.
                                         The --dry-run flag can be used in combination with the --output flag to
                                         view the Kubernetes resource(s) without sending anything to the server.
      --dry-run-with-image-upload      similar to --dry-run, but with container image uploads allowed.
                                         This flag is provided as a convenience for kp commands that can output Kubernetes
                                         resource with generated container image references. A "kubectl apply -f" of the
                                         resource from --output without image uploads will result in a reconcile failure.
  -f, --filename string                dependency descriptor filename
      --force                          import without confirmation when showing changes
  -h, --help                           help for import
      --output string                  print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                         The output can be used with the "kubectl apply -f" command. To allow this, the command
                                         updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
                                         The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
      --registry-ca-cert-path string   add CA certificate for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
      --show-changes                   show a summary of resource changes before importing
```

### SEE ALSO

* [kp](kp.md)	 - 

