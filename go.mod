module github.com/pivotal/build-service-cli

go 1.14

require (
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.5.1
	github.com/google/go-containerregistry v0.1.1
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/pivotal/kpack v0.0.10-0.20200730163525-f4aba175f503
	github.com/pkg/errors v0.9.1
	github.com/sclevine/spec v1.4.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.5
