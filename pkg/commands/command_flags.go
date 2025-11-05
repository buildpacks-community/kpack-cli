// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/buildpacks-community/kpack-cli/pkg/kpackcompat"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

const (
	caCertPathFlag  = "registry-ca-cert-path"
	verifyCertsFlag = "registry-verify-certs"

	caCertPathFlagUsage  = "add CA certificate for registry API (format: /tmp/ca.crt)"
	verifyCertsFlagUsage = "set whether to verify server's certificate chain and host name"
	dryRunUsage          = `perform validation with no side-effects; no objects are sent to the server.
  The --dry-run flag can be used in combination with the --output flag to
  view the Kubernetes resource(s) without sending anything to the server.`
	dryRunImgUploadUsage = `similar to --dry-run, but with container image uploads allowed.
  This flag is provided as a convenience for kp commands that can output Kubernetes
  resource with generated container image references. A "kubectl apply -f" of the
  resource from --output without image uploads will result in a reconcile failure.`
)

var outputUsage = fmt.Sprintf(`print Kubernetes resources in the specified format; supported formats are: yaml, json.
  The output can be used with the "kubectl apply -f" command. To allow this, the command
  updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.
  The APIVersion of the outputted resources will always be the latest APIVersion known to kp (currently: %s).`, kpackcompat.LatestKpackAPIVersion)

func SetTLSFlags(cmd *cobra.Command, cfg *registry.TLSConfig) {
	cmd.Flags().StringVar(&cfg.CaCertPath, caCertPathFlag, "", caCertPathFlagUsage)
	cmd.Flags().BoolVar(&cfg.VerifyCerts, verifyCertsFlag, true, verifyCertsFlagUsage)
}

func SetDryRunOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(DryRunFlag, false, dryRunUsage)
	cmd.Flags().String(OutputFlag, "", outputUsage)
}

func SetImgUploadDryRunOutputFlags(cmd *cobra.Command) {
	SetDryRunOutputFlags(cmd)
	cmd.Flags().Bool(DryRunImgUploadFlag, false, dryRunImgUploadUsage)
}
