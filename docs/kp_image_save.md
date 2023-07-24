## kp image save

Create or patch an image resource

### Synopsis

Create or patch an image resource by providing command line arguments.
This image resource will be created only if it does not exist in the provided namespace, otherwise it will be patched.

The --tag flag is required for a create but is immutable and will be ignored for a patch.
The --cache-size flag can only be used to create or increase the size of the existing cache.

The namespace defaults to the kubernetes current-context namespace.

The flags for this command determine how the build will retrieve source code:

  "--git" and "--git-revision" to use Git based source
  "--blob" to use source code hosted in a blob store
  "--local-path" to use source code from the local machine

Local source code will be pushed to the same registry provided for the image resource tag.
--local-path-destination-image can be used to specify the repository of the source code image.
If not specified, the source code image will be pushed to the <image-tag-repo>-source repo.
Therefore, you must have credentials to access the registry on your machine.

Environment variables may be provided by using the "--env" flag or deleted by using the "--delete-env" flag.
For each environment variable, supply the "--env" flag followed by the key value pair.
For example, "--env key1=value1 --env key2=value2 --delete-env key3".

Service bindings may be provided by using the "--service-binding" flag or deleted by using the "--delete-service-binding" flag.
For each service binding, supply the "--service-binding" flag followed by the <KIND>:<APIVERSION>:<NAME> or just <NAME> which will default the kind to "Secret".
For example, "--service-binding my-secret-1 --service-binding CustomProvisionedService:v1beta1:my-ps --delete-service-binding Secret:v1:my-secret-2"

Env vars can be used for registry auth as described in https://github.com/vmware-tanzu/kpack-cli/blob/main/docs/auth.md


```
kp image save <name> --tag <tag> [flags]
```

### Examples

```
kp image create my-image --tag my-registry.com/my-repo --git https://my-repo.com/my-app.git --git-revision my-branch
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob
kp image save my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code
kp image save my-image --tag my-registry.com/my-repo --local-path /path/to/local/source/code --builder my-builder -n my-namespace
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --env foo=bar --env color=red --env food=apple --delete-env apple --delete-env potato
kp image save my-image --tag my-registry.com/my-repo --blob https://my-blob-host.com/my-blob --service-binding my-secret --service-binding CustomProvisionedService:v1:my-ps --delete-service-binding Secret:v1:my-secret-2
```

### Options

```
      --additional-tag stringArray            additional tags to push the OCI image to
      --blob string                           source code blob url
  -b, --builder string                        builder name
      --cache-size string                     cache size as a kubernetes quantity (default "2G")
  -c, --cluster-builder string                cluster builder name
      --delete-additional-tag stringArray     additional tags to remove
  -d, --delete-env stringArray                build time environment variables to remove
      --delete-service-binding stringArray    build time service bindings to remove
      --dry-run                               perform validation with no side-effects; no objects are sent to the server.
                                                The --dry-run flag can be used in combination with the --output flag to
                                                view the Kubernetes resource(s) without sending anything to the server.
      --dry-run-with-image-upload             similar to --dry-run, but with container image uploads allowed.
                                                This flag is provided as a convenience for kp commands that can output Kubernetes
                                                resource with generated container image references. A "kubectl apply -f" of the
                                                resource from --output without image uploads will result in a reconcile failure.
  -e, --env stringArray                       build time environment variables
      --git string                            git repository url
      --git-revision string                   git revision such as commit, tag, or branch (default "main")
  -h, --help                                  help for save
      --local-path string                     path to local source code
      --local-path-destination-image string   registry location of source image
  -n, --namespace string                      kubernetes namespace
      --output string                         print Kubernetes resources in the specified format; supported formats are: yaml, json.
                                                The output can be used with the "kubectl apply -f" command. To allow this, the command
                                                updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
                                                The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: v1alpha2).
      --registry-ca-cert-path string          add CA certificate for registry API (format: /tmp/ca.crt)
      --registry-verify-certs                 set whether to verify server's certificate chain and host name (default true)
      --service-account string                service account name to use
  -s, --service-binding stringArray           build time service bindings to add/replace
      --sub-path string                       build code at the sub path located within the source code directory
  -t, --tag string                            registry location where the image will be created
  -w, --wait                                  wait for image create to be reconciled and tail resulting build logs
```

### SEE ALSO

* [kp image](kp_image.md)	 - Image commands

