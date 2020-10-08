## kp builder create

Create a builder

### Synopsis

Create a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.

```
kp builder create <name> --tag <tag> [flags]
```

### Examples

```
kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
```

### Options

```
      --dry-run            only print the object that would be sent, without sending it
  -h, --help               help for create
  -n, --namespace string   kubernetes namespace
  -o, --order string       path to buildpack order yaml
      --output string      output format. supported formats are: yaml, json
  -s, --stack string       stack resource to use (default "default")
      --store string       buildpack store to use (default "default")
  -t, --tag string         registry location where the builder will be created
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

