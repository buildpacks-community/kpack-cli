## kp builder status

Display status of a builder

### Synopsis

Prints detailed information about the status of a specific builder in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp builder status <name> [flags]
```

### Examples

```
kp builder status my-builder
kp builder status -n my-namespace other-builder
```

### Options

```
  -h, --help               help for status
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

