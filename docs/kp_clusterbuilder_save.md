## kp clusterbuilder save

Create or patch a cluster builder

### Synopsis

Create or patch a cluster builder by providing command line arguments.
The cluster builder will be created only if it does not exist, otherwise it is patched.

Tag when not specified, defaults to a combination of the canonical repository and specified builder name.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.

No defaults will be assumed for patches.


```
kp clusterbuilder save <name> [flags]
```

### Examples

```
kp cb save my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb save my-builder --order /path/to/order.yaml
kp cb save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp cb save my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
```

### Options

```
      --dry-run         only print the object that would be sent, without sending it
  -h, --help            help for save
  -o, --order string    path to buildpack order yaml
      --output string   output format. supported formats are: yaml, json
  -s, --stack string    stack resource to use (default "default" for a create)
      --store string    buildpack store to use (default "default" for a create)
  -t, --tag string      registry location where the builder will be created
```

### SEE ALSO

* [kp clusterbuilder](kp_clusterbuilder.md)	 - ClusterBuilder Commands

