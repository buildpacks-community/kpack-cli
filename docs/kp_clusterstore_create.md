## kp clusterstore create

Create a cluster store

### Synopsis

Create a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

This store will be created only if it does not exist.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.


```
kp clusterstore create <store> <buildpackage> [<buildpackage>...] [flags]
```

### Examples

```
kp store create my-store my-registry.com/my-buildpackage
kp clusterstore create my-store --buildpackage my-registry.com/my-buildpackage --buildpackage my-registry.com/my-other-buildpackage
kp clusterstore create my-store --buildpackage ../path/to/my-local-buildpackage.cnb
```

### Options

```
  -b, --buildpackage stringArray   location of the buildpackage
  -h, --help                       help for create
```

### SEE ALSO

* [kp clusterstore](kp_clusterstore.md)	 - Cluster Store Commands

