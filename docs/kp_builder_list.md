## kp builder list

List available builders

### Synopsis

Prints a table of the most important information about the available builders in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp builder list [flags]
```

### Examples

```
kp builder list
kp builder list -n my-namespace
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

