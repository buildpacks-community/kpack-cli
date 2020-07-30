## kp import

Import dependencies for stores, stacks, and cluster builders

### Synopsis

This operation will create or update stores, stacks, and cluster builders defined in the dependency descriptor.

```
kp import -f <filename> [flags]
```

### Examples

```
kp import -f dependencies.yaml
cat dependencies.yaml | kp import -f -
```

### Options

```
  -f, --filename string   dependency descriptor filename
  -h, --help              help for import
```

### SEE ALSO

* [kp](kp.md)	 - 

