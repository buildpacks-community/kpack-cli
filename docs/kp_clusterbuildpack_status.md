## kp clusterbuildpack status

Display status of a buildpack

### Synopsis

Prints detailed information about the status of a specific buildpack in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp clusterbuildpack status <name> [flags]
```

### Examples

```
kp clusterbuildpack status my-buildpack
```

### Options

```
  -h, --help               help for status
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp clusterbuildpack](kp_clusterbuildpack.md)	 - ClusterBuildpack Commands

