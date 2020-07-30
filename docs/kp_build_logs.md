## kp build logs

Tails logs for an image build

### Synopsis

Tails logs from the containers of a specific build of an image in the provided namespace.

build defaults to the latest build number.
namespace defaults to the kubernetes current-context namespace.

```
kp build logs <image-name> [flags]
```

### Examples

```
kp build logs my-image
kp build logs my-image -b 2 -n my-namespace
```

### Options

```
  -b, --build string       build number
  -h, --help               help for logs
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

