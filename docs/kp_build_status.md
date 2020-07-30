## kp build status

Display status for an image build

### Synopsis

Prints detailed information about the status of a specific build of an image in the provided namespace.

build defaults to the latest build number.
namespace defaults to the kubernetes current-context namespace.

```
kp build status <image-name> [flags]
```

### Examples

```
kp build status my-image
kp build status my-image -b 2 -n my-namespace
```

### Options

```
  -b, --build string       build number
  -h, --help               help for status
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

