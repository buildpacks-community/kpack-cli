## kp image list

List images

### Synopsis

Prints a table of the most important information about images in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp image list [flags]
```

### Examples

```
kp image list
kp image list -A
kp image list -n my-namespace
kp image list --filter ready=true --filter latest-reason=commit,trigger
```

### Options

```
  -A, --all-namespaces       Return objects found in all namespaces
      --filter stringArray   Each new filter argument requires an additional filter flag.
                             Multiple values can be provided using comma separation.
                             Supported filters and values:
                               builder=string
                               clusterbuilder=string
                               latest-reason=commit,trigger,config,stack,buildpack
                               ready=true,false,unknown
  -h, --help                 help for list
  -n, --namespace string     kubernetes namespace
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

