## kp config default-service-account

Set or Get the default service account

### Synopsis

Set or Get the default service account 

The default service account will be set as the service account on all cluster builders created with kp and the secrets on the service account will used to provide credentials to write cluster builder images.

This data is stored in a config map in the kpack namespace called kp-config. 
The kp-config config map also contains the default repository which is the location that imported and cluster-level resources are stored.

If this config map doesn't exist, it will automatically be created by running this command, but the default repository field will be empty.


```
kp config default-service-account [name] [flags]
```

### Examples

```
kp config default-service-account
kp config default-service-account my-service-account
kp config default-service-account my-service-account --service-account-namespace default
```

### Options

```
  -h, --help                               help for default-service-account
      --service-account-namespace string   namespace of default service account (default "kpack")
```

### SEE ALSO

* [kp config](kp_config.md)	 - Config commands

