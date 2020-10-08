## kp clusterstore delete

Delete a cluster store

### Synopsis

Delete a specific cluster-scoped buildpack store.

WARNING: Builders referring to buildpacks from this store will no longer schedule rebuilds for buildpack updates.

```
kp clusterstore delete <store> [flags]
```

### Examples

```
kp clusterstore delete my-store
```

### Options

```
  -f, --force   force deletion without confirmation
  -h, --help    help for delete
```

### SEE ALSO

* [kp clusterstore](kp_clusterstore.md)	 - ClusterStore Commands

