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
	"github.com/pivotal/build-service-cli/pkg/secret"
	"github.com/pivotal/build-service-cli/pkg/source"
	"github.com/pivotal/build-service-cli/pkg/stack"
)

var (
	Version   = "dev"
	CommitSHA = ""
)

func main() {
	var contextProvider commands.CommandContextProvider

	rootCmd := &cobra.Command{
		Use: "tbctl",
	}
	rootCmd.AddCommand(
		getVersionCommand(),
		getImageCommand(contextProvider),
		getBuildCommand(contextProvider),
		getSecretCommand(contextProvider),
		getClusterBuilderCommand(contextProvider),
		getBuilderCommand(contextProvider),
		getStackCommand(contextProvider),
		getStoreCommand(contextProvider),
	)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display tbctl version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version+" "+CommitSHA)
		},
	}
	return versionCmd
}

func getImageCommand(contextProvider commands.ContextProvider) *cobra.Command {
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
		imgcmds.NewCreateCommand(contextProvider, imageFactory),
		imgcmds.NewPatchCommand(contextProvider, imagePatchFactory),
		imgcmds.NewListCommand(contextProvider),
		imgcmds.NewDeleteCommand(contextProvider),
		imgcmds.NewTriggerCommand(contextProvider),
		imgcmds.NewStatusCommand(contextProvider),
	)
	return imageRootCmd
}

func getBuildCommand(contextProvider commands.ContextProvider) *cobra.Command {
	buildRootCmd := &cobra.Command{
		Use:   "build",
		Short: "Build Commands",
	}
	buildRootCmd.AddCommand(
		buildcmds.NewListCommand(contextProvider),
		buildcmds.NewStatusCommand(contextProvider),
		buildcmds.NewLogsCommand(contextProvider),
	)
	return buildRootCmd
}

func getSecretCommand(contextProvider commands.ContextProvider) *cobra.Command {
	credentialFetcher := &commands.CredentialFetcher{}
	secretFactory := &secret.Factory{
		CredentialFetcher: credentialFetcher,
	}

	secretRootCmd := &cobra.Command{
		Use:   "secret",
		Short: "Secret Commands",
	}
	secretRootCmd.AddCommand(
		secretcmds.NewCreateCommand(contextProvider, secretFactory),
		secretcmds.NewDeleteCommand(contextProvider),
		secretcmds.NewListCommand(contextProvider),
	)
	return secretRootCmd
}

func getClusterBuilderCommand(contextProvider commands.ContextProvider) *cobra.Command {
	clusterBuilderRootCmd := &cobra.Command{
		Use:     "custom-cluster-builder",
		Short:   "Custom Cluster Builder Commands",
		Aliases: []string{"ccb"},
	}
	clusterBuilderRootCmd.AddCommand(
		clusterbuildercmds.NewApplyCommand(contextProvider),
		clusterbuildercmds.NewListCommand(contextProvider),
		clusterbuildercmds.NewStatusCommand(contextProvider),
		clusterbuildercmds.NewDeleteCommand(contextProvider),
	)
	return clusterBuilderRootCmd
}

func getStackCommand(contextProvider commands.ContextProvider) *cobra.Command {
	stackFactory := &stack.Factory{
		Fetcher:   &image.Fetcher{},
		Relocator: &image.Relocator{},
	}

	stackRootCmd := &cobra.Command{
		Use:   "stack",
		Short: "Stack Commands",
	}
	stackRootCmd.AddCommand(
		stackcmds.NewCreateCommand(contextProvider, stackFactory),
		stackcmds.NewListCommand(contextProvider),
		stackcmds.NewStatusCommand(contextProvider),
		stackcmds.NewUpdateCommand(contextProvider, &image.Fetcher{}, &image.Relocator{}),
		stackcmds.NewDeleteCommand(contextProvider),
	)
	return stackRootCmd
}

func getBuilderCommand(contextProvider commands.ContextProvider) *cobra.Command {
	builderRootCmd := &cobra.Command{
		Use:     "custom-builder",
		Short:   "Custom Builder Commands",
		Aliases: []string{"cb"},
	}
	builderRootCmd.AddCommand(
		buildercmds.NewApplyCommand(contextProvider),
		buildercmds.NewListCommand(contextProvider),
		buildercmds.NewDeleteCommand(contextProvider),
		buildercmds.NewStatusCommand(contextProvider),
	)
	return builderRootCmd
}

func getStoreCommand(contextProvider commands.ContextProvider) *cobra.Command {
	bpUploader := &buildpackage.Uploader{
		Fetcher:   &image.Fetcher{},
		Relocator: &image.Relocator{},
	}

	storeRootCommand := &cobra.Command{
		Use:   "store",
		Short: "Store Commands",
	}
	storeRootCommand.AddCommand(
		store.NewAddCommand(contextProvider, bpUploader),
		store.NewStatusCommand(contextProvider),
		store.NewDeleteCommand(contextProvider),
	)

	return storeRootCommand
}
