## kp buildpack list

List available buildpacks

### Synopsis

Prints a table of the most important information about the available buildpacks in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp buildpack list [flags]
```

### Examples

```
kp buildpack list
kp buildpack list -n my-namespace
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp buildpack](kp_buildpack.md)	 - Buildpack Commands

