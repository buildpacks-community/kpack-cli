## kp config default-repository

Set or Get the default repository

### Synopsis

Set or Get the default repository 

The default repository is the location where imported and cluster-level resources are stored. It is required to be configured to use kp.

This data is stored in a config map in the kpack namespace called kp-config. 
The kp-config config map also contains a service account that contains the secrets required to write to the default repository.

If this config map doesn't exist, it will automatically be created by running this command, using the default service account in the kpack namespace as the default service account.


```
kp config default-repository [url] [flags]
```

### Examples

```
kp config default-repository
kp config default-repository my-registry.com/my-default-repo
```

### Options

```
  -h, --help   help for default-repository
```

### SEE ALSO

* [kp config](kp_config.md)	 - Config commands

