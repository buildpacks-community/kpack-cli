## kp clusterlifecycle create

Create a clusterlifecycle

### Synopsis

Create a cluster-scoped lifecycle by providing command line arguments.

The image will be uploaded to the default repository.
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.

Env vars can be used for registry auth as described in https://github.com/buildpacks-community/kpack-cli/blob/main/docs/auth.md

The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
The default service account used is read from the "default.repository.serviceaccount" key in the "kp-config" ConfigMap within "kpack" namespace.


```
kp clusterlifecycle create <name> [flags]
```

### Examples

```
kp clusterlifecycle create my-lifecycle --image buildpacksio/lifecycle

```

### Options

```
      --dry-run                        perform validation with no side-effects; no objects are sent to the server.
                                         The --dry-run flag can be used in combination with the --output flag to
                                         view the Kubernetes resource(s) without sending anything to the server.
      --dry-run-with-image-upload      similar to --dry-run, but with container image uploads allowed.
                                         This flag is provided as a convenience for kp commands that can output Kubernetes
                                         resource with generated container image references. A "kubectl apply -f" of the
                                         resource from --output without image uploads will result in a reconcile failure.
  -h, --help                           help for create
  -i, --image string                   image tag or local tar file path
      --output string                  print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                         The output can be used with the "kubectl apply -f" command. To allow this, the command
                                         updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
                                         The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
      --registry-ca-cert-path string   add CA certificate for registry API (format: /tmp/ca.crt)
      --registry-verify-certs          set whether to verify server's certificate chain and host name (default true)
```

### SEE ALSO

* [kp clusterlifecycle](kp_clusterlifecycle.md)	 - ClusterLifecycle Commands

