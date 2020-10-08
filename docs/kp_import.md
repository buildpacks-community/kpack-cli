## kp import

Import dependencies for stores, stacks, and cluster builders

### Synopsis

This operation will create or update stores, stacks, and cluster builders defined in the dependency descriptor.

```
kp import -f <filename> [flags]
```

### Examples

```
kp import -f dependencies.yaml
cat dependencies.yaml | kp import -f -
```

### Options

```
      --dry-run                        only print the object that would be sent, without sending it
  -f, --filename string                dependency descriptor filename
  -h, --help                           help for import
      --output string                  output format. supported formats are: yaml, json
      --registry-ca-cert-path string   add CA certificates for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
```

### SEE ALSO

* [kp](kp.md)	 - 

