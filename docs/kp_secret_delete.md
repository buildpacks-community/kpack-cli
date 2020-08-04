## kp secret delete

Delete secret

### Synopsis

Deletes a specific secret in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp secret delete <name> [flags]
```

### Examples

```
kp secret delete my-secret
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   kubernetes namespace
```

### SEE ALSO

* [kp secret](kp_secret.md)	 - Secret Commands

