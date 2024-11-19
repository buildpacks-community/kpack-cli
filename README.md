# kp cli to interact with [kpack](https://github.com/pivotal/kpack)

## Installation

`amd64` and `arm64` binaries for both Linux and OSX can be installed through [Homebrew](https://github.com/buildpacks-community/homebrew-kpack-cli)

```
brew tap buildpacks-community/kpack-cli
brew install kp
```

You can also download the binary directly from [releases](https://github.com/buildpacks-community/kpack-cli/releases)

## Contributing

Please read the [contributing](CONTRIBUTING.md) doc to begin contributing.

## Building Locally

```
go install ./cmd/kp
```

## Getting Started

```
kp --help
```

## [Documentation](docs/kp.md)

Update docs with `go run cmd/docs/main.go`

## [Docker Image](https://hub.docker.com/r/kpack/kp) (kpack/kp)
