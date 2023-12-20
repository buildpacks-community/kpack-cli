## kp build logs

Tails logs for an image resource or for a build resource

### Synopsis

Tails logs from the containers of a specific build of an image or build resource in the provided namespace.

By default command will assume user provided an Image name and will attempt to find builds associated with that Image.
If no builds are found matching the Image name, It will assume the provided argument was a Build name.

The build defaults to the latest build number.
The namespace defaults to the kubernetes current-context namespace.

Use the flag --timestamps to include the timestamps for the logs

```
kp build logs <image-name|build-name> [flags]
```

### Examples

```
kp build logs my-image
kp build logs my-image -b 2 -n my-namespace
kp build logs my-build-name -n my-namespace
```

### Options

```
  -b, --build string       build number
  -h, --help               help for logs
  -n, --namespace string   kubernetes namespace
  -t, --timestamps         show log timestamps
```

### SEE ALSO

* [kp build](kp_build.md)	 - Build Commands

