## kp clusterbuilder patch

Patch an existing cluster builder configuration

### Synopsis

Patch an existing clusterbuilder configuration by providing command line arguments.

A buildpack order must be provided with either the path to an order yaml or via the --buildpack flag.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group.

```
kp clusterbuilder patch <name> [flags]
```

### Examples

```
kp cb patch my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb patch my-builder --order /path/to/order.yaml
kp cb patch my-builder --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1
```

### Options

```
  -b, --buildpack strings   buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'
                              repeat for each buildpack in order, or supply once with comma-separated list
      --dry-run             perform validation with no side-effects; no objects are sent to the server.
                              The --dry-run flag can be used in combination with the --output flag to
                              view the Kubernetes resource(s) without sending anything to the server.
  -h, --help                help for patch
  -o, --order string        path to buildpack order yaml
      --output string       print Kubernetes resources in the specified format; supported formats are: yaml, json.
                              The output can be used with the "kubectl apply -f" command. To allow this, the command 
                              updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
  -s, --stack string        stack resource to use
      --store string        buildpack store to use
  -t, --tag string          registry location where the builder will be created
```

### SEE ALSO

* [kp clusterbuilder](kp_clusterbuilder.md)	 - ClusterBuilder Commands

