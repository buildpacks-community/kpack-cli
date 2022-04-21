## kp image delete

Delete an image resource

### Synopsis

Delete an image resource and its associated builds in the provided namespace.

namespace defaults to the kubernetes current-context namespace.
this will not delete your OCI image in the registry

```
kp image delete <name> [flags]
```

### Examples

```
kp image delete my-image
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

