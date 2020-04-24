package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/commands"
	buildcmds "github.com/pivotal/build-service-cli/pkg/commands/build"
	buildercmds "github.com/pivotal/build-service-cli/pkg/commands/builder"
	clusterbuildercmds "github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
	"github.com/pivotal/build-service-cli/pkg/secret"
)

var Version = "dev"

func main() {
	defaultNamespace, err := k8s.GetDefaultNamespace()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	kpackClient, err := k8s.NewKpackClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	k8sClient, err := k8s.NewK8sClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	buildRootCmd := &cobra.Command{
		Use:   "build",
		Short: "Build Commands",
	}
	buildRootCmd.AddCommand(
		buildcmds.NewListCommand(kpackClient, defaultNamespace),
		buildcmds.NewStatusCommand(kpackClient, defaultNamespace),
		buildcmds.NewLogsCommand(kpackClient, k8sClient, defaultNamespace),
	)

	sourceUploader := &registry.SourceUploader{}

	imageFactory := &image.Factory{
		SourceUploader: sourceUploader,
	}

	imageRootCmd := &cobra.Command{
		Use:   "image",
		Short: "Image commands",
	}
	imageRootCmd.AddCommand(
		imgcmds.NewGetCommand(kpackClient, defaultNamespace),
		imgcmds.NewApplyCommand(kpackClient, defaultNamespace),
		imgcmds.NewListCommand(kpackClient, defaultNamespace),
		imgcmds.NewDeleteCommand(kpackClient, defaultNamespace),
		imgcmds.NewTriggerCommand(kpackClient, defaultNamespace),
		imgcmds.NewStatusCommand(kpackClient, defaultNamespace),
		imgcmds.NewCreateCommand(kpackClient, imageFactory, defaultNamespace),
	)

	credentialFetcher := &commands.CredentialFetcher{}

	secretFactory := &secret.Factory{
		CredentialFetcher: credentialFetcher,
	}

	secretRootCmd := &cobra.Command{
		Use:   "secret",
		Short: "Secret Commands",
	}
	secretRootCmd.AddCommand(
		secretcmds.NewCreateCommand(k8sClient, secretFactory, defaultNamespace),
		secretcmds.NewDeleteCommand(k8sClient, defaultNamespace),
		secretcmds.NewListCommand(k8sClient, defaultNamespace),
	)

	clusterBuilderRootCmd := &cobra.Command{
		Use:   "clusterbuilder",
		Short: "Cluster Builder Commands",
	}
	clusterBuilderRootCmd.AddCommand(
		clusterbuildercmds.NewApplyCommand(kpackClient),
		clusterbuildercmds.NewListCommand(kpackClient),
		clusterbuildercmds.NewStatusCommand(kpackClient),
		clusterbuildercmds.NewDeleteCommand(kpackClient),
	)

	builderRootCmd := &cobra.Command{
		Use:   "builder",
		Short: "Builder Commands",
	}
	builderRootCmd.AddCommand(
		buildercmds.NewApplyCommand(kpackClient, defaultNamespace),
	)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display tbctl version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version)
		},
	}

	rootCmd := &cobra.Command{
		Use: "tbctl",
	}
	rootCmd.AddCommand(
		versionCmd,
		imageRootCmd,
		buildRootCmd,
		secretRootCmd,
		clusterBuilderRootCmd,
		builderRootCmd,
	)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
