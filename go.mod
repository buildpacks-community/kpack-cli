module github.com/vmware-tanzu/kpack-cli

go 1.16

require (
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.6.0
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/pivotal/kpack v0.3.2-0.20211004222222-a14c908acace
	github.com/pkg/errors v0.9.1
	github.com/sclevine/spec v1.4.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	knative.dev/pkg v0.0.0-20210902173607-844a6bc45596
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/prometheus/common => github.com/prometheus/common v0.26.0
	k8s.io/api => k8s.io/api v0.20.11
	k8s.io/client-go => k8s.io/client-go v0.20.11
)
