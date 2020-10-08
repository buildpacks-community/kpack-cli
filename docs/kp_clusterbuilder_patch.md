## kp clusterbuilder patch

Patch an existing cluster builder configuration

### Synopsis

 

```
kp clusterbuilder patch <name> [flags]
```

### Examples

```
kp cb patch my-builder
```

### Options

```
      --dry-run         only print the object that would be sent, without sending it
  -h, --help            help for patch
  -o, --order string    path to buildpack order yaml
      --output string   output format. supported formats are: yaml, json
  -s, --stack string    stack resource to use
      --store string    buildpack store to use
  -t, --tag string      registry location where the builder will be created
```

### SEE ALSO

* [kp clusterbuilder](kp_clusterbuilder.md)	 - ClusterBuilder Commands

