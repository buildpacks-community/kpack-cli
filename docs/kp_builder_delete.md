## kp builder delete

Delete a builder

### Synopsis

Delete a builder in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp builder delete <name> [flags]
```

### Examples

```
kp builder delete my-builder
kp builder delete -n my-namespace other-builder
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

