## kp clusterbuilder create

Create a cluster builder

### Synopsis

Create a cluster builder by providing command line arguments.
The cluster builder will be created only if it does not exist.

Tag when not specified, defaults to a combination of the canonical repository and specified builder name.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.


```
kp clusterbuilder create <name> [flags]
```

### Examples

```
kp cb create my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb create my-builder --order /path/to/order.yaml
kp cb create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp cb create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
```

### Options

```
  -h, --help           help for create
  -o, --order string   path to buildpack order yaml
  -s, --stack string   stack resource to use (default "default")
      --store string   buildpack store to use (default "default")
  -t, --tag string     registry location where the builder will be created
```

### SEE ALSO

* [kp clusterbuilder](kp_clusterbuilder.md)	 - Cluster Builder Commands

