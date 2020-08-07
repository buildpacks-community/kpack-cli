## kp clusterstack create

Create a cluster stack

### Synopsis

Create a cluster-scoped stack by providing command line arguments.

The run and build images will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.
Additionally, your cluster must have read access to the registry.

The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.


```
kp clusterstack create <name> [flags]
```

### Examples

```
kp clusterstack create my-stack --build-image my-registry.com/build --run-image my-registry.com/run
kp clusterstack create my-stack --build-image ../path/to/build.tar --run-image ../path/to/run.tar
```

### Options

```
  -b, --build-image string   build image tag or local tar file path
  -h, --help                 help for create
  -r, --run-image string     run image tag or local tar file path
```

### SEE ALSO

* [kp clusterstack](kp_clusterstack.md)	 - Cluster Stack Commands

