module github.com/vmware-tanzu/kpack-cli

go 1.14

require (
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.6.0
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/pivotal/kpack v0.3.1
	github.com/pkg/errors v0.9.1
	github.com/sclevine/spec v1.4.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	knative.dev/pkg v0.0.0-20210819054404-bda81c029160
	sigs.k8s.io/yaml v1.2.0
)
