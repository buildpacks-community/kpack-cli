## kp image trigger

Trigger an image build

### Synopsis

Trigger a build using current inputs for a specific image in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp image trigger <name> [flags]
```

### Examples

```
kp image trigger my-image
```

### Options

```
  -h, --help               help for trigger
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

