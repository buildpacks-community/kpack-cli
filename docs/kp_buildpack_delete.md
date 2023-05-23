## kp buildpack delete

Delete a buildpack

### Synopsis

Delete a buildpack in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp buildpack delete <name> [flags]
```

### Examples

```
kp buildpack delete my-buildpack
kp buildpack delete -n my-namespace other-buildpack
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp buildpack](kp_buildpack.md)	 - Buildpack Commands

