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
	var cmdContext commands.CommandContext

	rootCmd := &cobra.Command{
		Use: "tbctl",
	}
	rootCmd.AddCommand(
		getVersionCommand(),
		getImageCommand(cmdContext),
		getBuildCommand(cmdContext),
		getSecretCommand(cmdContext),
		getClusterBuilderCommand(cmdContext),
		getBuilderCommand(cmdContext),
		getStackCommand(cmdContext),
		getStoreCommand(cmdContext),
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

func getImageCommand(cmdContext commands.CommandContext) *cobra.Command {
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
		imgcmds.NewCreateCommand(cmdContext, imageFactory),
		imgcmds.NewPatchCommand(cmdContext, imagePatchFactory),
		imgcmds.NewListCommand(cmdContext),
		imgcmds.NewDeleteCommand(cmdContext),
		imgcmds.NewTriggerCommand(cmdContext),
		imgcmds.NewStatusCommand(cmdContext),
	)
	return imageRootCmd
}

func getBuildCommand(cmdContext commands.CommandContext) *cobra.Command {
	buildRootCmd := &cobra.Command{
		Use:   "build",
		Short: "Build Commands",
	}
	buildRootCmd.AddCommand(
		buildcmds.NewListCommand(cmdContext),
		buildcmds.NewStatusCommand(cmdContext),
		buildcmds.NewLogsCommand(cmdContext),
	)
	return buildRootCmd
}

func getSecretCommand(cmdContext commands.CommandContext) *cobra.Command {
	credentialFetcher := &commands.CredentialFetcher{}
	secretFactory := &secret.Factory{
		CredentialFetcher: credentialFetcher,
	}

	secretRootCmd := &cobra.Command{
		Use:   "secret",
		Short: "Secret Commands",
	}
	secretRootCmd.AddCommand(
		secretcmds.NewCreateCommand(cmdContext, secretFactory),
		secretcmds.NewDeleteCommand(cmdContext),
		secretcmds.NewListCommand(cmdContext),
	)
	return secretRootCmd
}

func getClusterBuilderCommand(cmdContext commands.CommandContext) *cobra.Command {
	clusterBuilderRootCmd := &cobra.Command{
		Use:     "custom-cluster-builder",
		Short:   "Custom Cluster Builder Commands",
		Aliases: []string{"ccb"},
	}
	clusterBuilderRootCmd.AddCommand(
		clusterbuildercmds.NewApplyCommand(cmdContext),
		clusterbuildercmds.NewListCommand(cmdContext),
		clusterbuildercmds.NewStatusCommand(cmdContext),
		clusterbuildercmds.NewDeleteCommand(cmdContext),
	)
	return clusterBuilderRootCmd
}

func getStackCommand(cmdContext commands.CommandContext) *cobra.Command {
	stackFactory := &stack.Factory{
		Fetcher:   &image.Fetcher{},
		Relocator: &image.Relocator{},
	}

	stackRootCmd := &cobra.Command{
		Use:   "stack",
		Short: "Stack Commands",
	}
	stackRootCmd.AddCommand(
		stackcmds.NewCreateCommand(cmdContext, stackFactory),
		stackcmds.NewListCommand(cmdContext),
		stackcmds.NewStatusCommand(cmdContext),
		stackcmds.NewUpdateCommand(cmdContext, &image.Fetcher{}, &image.Relocator{}),
		stackcmds.NewDeleteCommand(cmdContext),
	)
	return stackRootCmd
}

func getBuilderCommand(cmdContext commands.CommandContext) *cobra.Command {
	builderRootCmd := &cobra.Command{
		Use:     "custom-builder",
		Short:   "Custom Builder Commands",
		Aliases: []string{"cb"},
	}
	builderRootCmd.AddCommand(
		buildercmds.NewApplyCommand(cmdContext),
		buildercmds.NewListCommand(cmdContext),
		buildercmds.NewDeleteCommand(cmdContext),
		buildercmds.NewStatusCommand(cmdContext),
	)
	return builderRootCmd
}

func getStoreCommand(cmdContext commands.CommandContext) *cobra.Command {
	bpUploader := &buildpackage.Uploader{
		Fetcher:   &image.Fetcher{},
		Relocator: &image.Relocator{},
	}

	storeRootCommand := &cobra.Command{
		Use:   "store",
		Short: "Store Commands",
	}
	storeRootCommand.AddCommand(
		store.NewAddCommand(cmdContext, bpUploader),
		store.NewStatusCommand(cmdContext),
		store.NewDeleteCommand(cmdContext),
	)

	return storeRootCommand
}
