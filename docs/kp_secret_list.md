## kp secret list

List secrets

### Synopsis

Prints a table of the most important information about secrets in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp secret list [flags]
```

### Examples

```
kp secret list
kp secret list -n my-namespace
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp secret](kp_secret.md)	 - Secret Commands

