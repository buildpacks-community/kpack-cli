package commands

import (
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

func SetTLSFlags(cmd *cobra.Command, cfg *registry.TLSConfig) {
	cmd.Flags().StringVar(&cfg.CaCertPath, "registry-ca-cert-path", "", "add CA certificates for registry API (format: /tmp/ca.crt)")
	cmd.Flags().BoolVar(&cfg.VerifyCerts, "registry-verify-certs", true, "set whether to verify server's certificate chain and host name")
}

func SetDryRunOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, "only print the object that would be sent, without sending it")
	cmd.Flags().String("output", "", "output format. supported formats are: yaml, json")
}
