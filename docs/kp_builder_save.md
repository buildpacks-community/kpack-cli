## kp builder save

Create or patch a builder

### Synopsis

Create or patch a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace, otherwise it will be patched.

The --tag flag is required for a create but is immutable and will be ignored for a patch.

No defaults will be assumed for patches.

The namespace defaults to the kubernetes current-context namespace.

```
kp builder save <name> [flags]
```

### Examples

```
kp builder save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
```

### Options

```
      --dry-run            only print the object that would be sent, without sending it
  -h, --help               help for save
  -n, --namespace string   kubernetes namespace
  -o, --order string       path to buildpack order yaml
      --output string      output format. supported formats are: yaml, json
  -s, --stack string       stack resource to use (default "default" for a create)
      --store string       buildpack store to use (default "default" for a create)
  -t, --tag string         registry location where the builder will be created
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

