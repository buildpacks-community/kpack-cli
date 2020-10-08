## kp builder patch

Patch an existing builder configuration

### Synopsis

 

```
kp builder patch <name> [flags]
```

### Examples

```
kp builder patch my-builder
```

### Options

```
      --dry-run            only print the object that would be sent, without sending it
  -h, --help               help for patch
  -n, --namespace string   kubernetes namespace
  -o, --order string       path to buildpack order yaml
      --output string      output format. supported formats are: yaml, json
  -s, --stack string       stack resource to use
      --store string       buildpack store to use
  -t, --tag string         registry location where the builder will be created
```

### SEE ALSO

* [kp builder](kp_builder.md)	 - Builder Commands

