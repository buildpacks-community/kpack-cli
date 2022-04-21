## kp image status

Display status of an image resource

### Synopsis

Prints detailed information about the status of a specific image resource in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp image status <name> [flags]
```

### Examples

```
kp image status my-image
kp image status my-other-image -n my-namespace
```

### Options

```
  -h, --help               help for status
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

