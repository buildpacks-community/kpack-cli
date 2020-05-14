package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/buildpackage"
	"github.com/pivotal/build-service-cli/pkg/commands"
	buildcmds "github.com/pivotal/build-service-cli/pkg/commands/build"
	buildercmds "github.com/pivotal/build-service-cli/pkg/commands/custombuilder"
	clusterbuildercmds "github.com/pivotal/build-service-cli/pkg/commands/customclusterbuilder"
	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	stackcmds "github.com/pivotal/build-service-cli/pkg/commands/stack"
	"github.com/pivotal/build-service-cli/pkg/commands/store"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/secret"
	"github.com/pivotal/build-service-cli/pkg/source"
	"github.com/pivotal/build-service-cli/pkg/stack"
)

var (
	Version   = "dev"
	CommitSHA = ""
)

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

	sourceUploader := &source.Uploader{}

	imageFactory := &image.Factory{
		SourceUploader: sourceUploader,
	}

	imagePatchFactory := &image.PatchFactory{
		SourceUploader: sourceUploader,
	}

	imageRootCmd := &cobra.Command{
		Use:   "image",
		Short: "Image commands",
	}
	imageRootCmd.AddCommand(
		imgcmds.NewCreateCommand(kpackClient, imageFactory, defaultNamespace),
		imgcmds.NewPatchCommand(kpackClient, imagePatchFactory, defaultNamespace),
		imgcmds.NewListCommand(kpackClient, defaultNamespace),
		imgcmds.NewDeleteCommand(kpackClient, defaultNamespace),
		imgcmds.NewTriggerCommand(kpackClient, defaultNamespace),
		imgcmds.NewStatusCommand(kpackClient, defaultNamespace),
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
		Use:     "custom-cluster-builder",
		Short:   "Custom Cluster Builder Commands",
		Aliases: []string{"ccb"},
	}
	clusterBuilderRootCmd.AddCommand(
		clusterbuildercmds.NewApplyCommand(kpackClient),
		clusterbuildercmds.NewListCommand(kpackClient),
		clusterbuildercmds.NewStatusCommand(kpackClient),
		clusterbuildercmds.NewDeleteCommand(kpackClient),
	)

	builderRootCmd := &cobra.Command{
		Use:     "custom-builder",
		Short:   "Custom Builder Commands",
		Aliases: []string{"cb"},
	}
	builderRootCmd.AddCommand(
		buildercmds.NewApplyCommand(kpackClient, defaultNamespace),
		buildercmds.NewListCommand(kpackClient, defaultNamespace),
		buildercmds.NewDeleteCommand(kpackClient, defaultNamespace),
		buildercmds.NewStatusCommand(kpackClient, defaultNamespace),
	)

	bpUploader := &buildpackage.Uploader{
		Fetcher:   &image.Fetcher{},
		Relocator: &image.Relocator{},
	}

	storeRootCommand := &cobra.Command{
		Use:   "store",
		Short: "Store Commands",
	}
	storeRootCommand.AddCommand(
		store.NewAddCommand(kpackClient, bpUploader),
		store.NewStatusCommand(kpackClient),
		store.NewDeleteCommand(kpackClient),
	)

	stackFactory := &stack.Factory{
		Fetcher:   &image.Fetcher{},
		Relocator: &image.Relocator{},
	}

	stackRootCmd := &cobra.Command{
		Use:   "stack",
		Short: "Stack Commands",
	}
	stackRootCmd.AddCommand(
		stackcmds.NewCreateCommand(kpackClient, stackFactory),
		stackcmds.NewListCommand(kpackClient),
		stackcmds.NewStatusCommand(kpackClient),
		stackcmds.NewUpdateCommand(kpackClient, &image.Fetcher{}, &image.Relocator{}),
		stackcmds.NewDeleteCommand(kpackClient),
	)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display tbctl version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version+" "+CommitSHA)
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
		stackRootCmd,
		storeRootCommand,
	)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
