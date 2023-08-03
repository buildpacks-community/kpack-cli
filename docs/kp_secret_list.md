## kp secret list

List secrets attached to a service account

### Synopsis

List secrets for a service account in the provided namespace.

A secret attached to a service account that does not exist in the specified namespace will be listed as AVAILABLE "false".

The namespace defaults to the kubernetes current-context namespace.

The service account defaults to "default".

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
  -h, --help                     help for list
  -n, --namespace string         kubernetes namespace
      --service-account string   service account to list secrets for (default "default")
```

### SEE ALSO

* [kp secret](kp_secret.md)	 - Secret Commands

