## kp build list

List builds for an image

### Synopsis

Prints a table of the most important information about builds for an image in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp build list <image-name> [flags]
```

### Examples

```
kp build list my-image
kp build list my-image -n my-namespace
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

