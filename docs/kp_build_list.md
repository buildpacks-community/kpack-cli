## kp build list

List builds

### Synopsis

Prints a table of the most important information about builds in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp build list [image-resource-name] [flags]
```

### Examples

```
kp build list
kp build list my-image
kp build list my-image -n my-namespace
kp build list -A
```

### Options

```
  -A, --all-namespaces     Return objects found in all namespaces
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

