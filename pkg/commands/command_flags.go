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
	cmd.Flags().Bool(DryRunFlag, false, `perform validation with no side-effects; no objects are sent to the server.
The --dry-run flag can be used in combination with the --output flag to
view the Kubernetes resource(s) without sending anything to the server.`)
	cmd.Flags().String(OutputFlag, "", `print Kubernetes resources in the specified format; supported formats are: yaml, json.
The output can be used with the "kubectl apply -f" command. To allow this, the command 
updates are redirected to stderr and only the Kubernetes resource(s) are written to stdout.`)
}

func SetImgUploadDryRunOutputFlags(cmd *cobra.Command) {
	SetDryRunOutputFlags(cmd)
	cmd.Flags().Bool(DryRunImgUploadFlag, false, `similar to --dry-run, but with container image uploads allowed.
This flag is provided as a convenience for kp commands that can output Kubernetes
resource with generated container image references. A "kubectl apply -f" of the
resource from --output without image uploads will result in a reconcile failure.`)
}
