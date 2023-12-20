## kp build status

Display status for an image resource or for a build resource

### Synopsis

Prints detailed information about the status of a specific build of an image resource or build resource in the provided namespace.

By default command will assume user provided an Image name and will attempt to find builds associated with that Image.
If no builds are found matching the Image name, It will assume the provided argument was a Build name.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.

```
kp build status <image-name|build-name> [flags]
```

### Examples

```
kp build status my-image
kp build status my-image -b 2 -n my-namespace
kp build status my-build-name -n my-namespace
```

### Options

```
  -b, --build string       build number
  -h, --help               help for status
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

