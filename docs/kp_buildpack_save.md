## kp buildpack save

Create or patch a buildpack

### Synopsis

Create or patch a buildpack by providing command line arguments.
The buildpack will be created only if it does not exist in the provided namespace, otherwise it will be patched.

No defaults will be assumed for patches.

The namespace defaults to the kubernetes current-context namespace.

```
kp buildpack save <name> [flags]
```

### Examples

```
kp buildpack save my-buildpack --image gcr.io/paketo-buildpacks/java
kp buildpack save my-buildpack --service-account my-other-sa
```

### Options

```
      --dry-run                  perform validation with no side-effects; no objects are sent to the server.
                                   The --dry-run flag can be used in combination with the --output flag to
                                   view the Kubernetes resource(s) without sending anything to the server.
  -h, --help                     help for save
  -i, --image string             registry location where the buildpack is located
  -n, --namespace string         kubernetes namespace
      --output string            print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                   The output can be used with the "kubectl apply -f" command. To allow this, the command
                                   updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
                                   The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
      --service-account string   service account name to use
```

### SEE ALSO

* [kp buildpack](kp_buildpack.md)	 - Buildpack Commands

