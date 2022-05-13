## kp builder create

Create a builder

### Synopsis

Create a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace.

A buildpack order must be provided with either the path to an order yaml or via the --buildpack flag.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group. 

The namespace defaults to the kubernetes current-context namespace.

```
kp builder create <name> --tag <tag> [flags]
```

### Examples

```
kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
kp builder create my-builder --tag my-registry.com/my-builder-tag --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1
```

### Options

```
  -b, --buildpack strings   buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'
                              repeat for each buildpack in order, or supply once with comma-separated list
      --dry-run             perform validation with no side-effects; no objects are sent to the server.
                              The --dry-run flag can be used in combination with the --output flag to
                              view the Kubernetes resource(s) without sending anything to the server.
  -h, --help                help for create
  -n, --namespace string    kubernetes namespace
  -o, --order string        path to buildpack order yaml
      --output string       print Kubernetes resources in the specified format; supported formats are: yaml, json.
                              The output can be used with the "kubectl apply -f" command. To allow this, the command
                              updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
                              The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
  -s, --stack string        stack resource to use (default "default")
      --store string        buildpack store to use (default "default")
  -t, --tag string          registry location where the builder will be created
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

