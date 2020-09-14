package commands

import (
	"github.com/pivotal/build-service-cli/pkg/registry"
	"github.com/spf13/cobra"
)

func SetTLSFlags(cmd *cobra.Command, cfg *registry.TLSConfig) {
	cmd.Flags().StringVar(&cfg.CaCertPath, "registry-ca-cert-path", "", "add CA certificates for registry API (format: /tmp/ca.crt)")
	cmd.Flags().BoolVar(&cfg.VerifyCerts, "registry-verify-certs", true, "set whether to verify server's certificate chain and host name (default true)")
}
