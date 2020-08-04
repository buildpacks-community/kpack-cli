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
kp image list -n my-namespace
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

