// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal/kpack/pkg/logs"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/buildpackage"
	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	buildcmds "github.com/pivotal/build-service-cli/pkg/commands/build"
	buildercmds "github.com/pivotal/build-service-cli/pkg/commands/builder"
	clusterbuildercmds "github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	clusterstackcmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	storecmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	importcmds "github.com/pivotal/build-service-cli/pkg/commands/import"
	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/image"
	importpkg "github.com/pivotal/build-service-cli/pkg/import"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
	"github.com/pivotal/build-service-cli/pkg/secret"
)

var (
	Version   = "dev"
	CommitSHA = ""
)

func main() {
	log.SetOutput(ioutil.Discard)

	var clientSetProvider k8s.DefaultClientSetProvider

	rootCmd := &cobra.Command{
		Use: "kp",
		Long: `kp controls the kpack installation on Kubernetes.

kpack extends Kubernetes and utilizes unprivileged kubernetes primitives to provide 
builds of OCI images as a platform implementation of Cloud Native Buildpacks (CNB).
Learn more about kpack @ https://github.com/pivotal/kpack`,
	}
	rootCmd.AddCommand(
		getVersionCommand(),
		getImageCommand(clientSetProvider),
		getBuildCommand(clientSetProvider),
		getSecretCommand(clientSetProvider),
		getClusterBuilderCommand(clientSetProvider),
		getBuilderCommand(clientSetProvider),
		getStackCommand(clientSetProvider),
		getStoreCommand(clientSetProvider),
		getImportCommand(clientSetProvider),
		getCompletionCommand(),
	)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	/* Generate Documentation /
	rootCmd.DisableAutoGenTag = true
	err := doc.GenMarkdownTree(rootCmd, "./docs")
	if err != nil {
		os.Exit(1)
	} /**/
}

func getVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display kp version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version+" "+CommitSHA)
		},
	}
	return versionCmd
}

func getImageCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	newImageWaiter := func(clientSet k8s.ClientSet) imgcmds.ImageWaiter {
		return logs.NewImageWaiter(clientSet.KpackClient, logs.NewBuildLogsClient(clientSet.K8sClient))
	}

	factory := &image.Factory{}

	imageRootCmd := &cobra.Command{
		Use:     "image",
		Short:   "Image commands",
		Aliases: []string{"images", "imgs", "img"},
	}
	imageRootCmd.AddCommand(
		imgcmds.NewCreateCommand(clientSetProvider, factory, newImageWaiter),
		imgcmds.NewPatchCommand(clientSetProvider, factory, newImageWaiter),
		imgcmds.NewSaveCommand(clientSetProvider, factory, newImageWaiter),
		imgcmds.NewListCommand(clientSetProvider),
		imgcmds.NewDeleteCommand(clientSetProvider),
		imgcmds.NewTriggerCommand(clientSetProvider),
		imgcmds.NewStatusCommand(clientSetProvider),
	)
	imageRootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return configureImageFactory(cmd, factory)
	}

	return imageRootCmd
}

func getBuildCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	buildRootCmd := &cobra.Command{
		Use:     "build",
		Short:   "Build Commands",
		Aliases: []string{"builds", "blds", "bld"},
	}
	buildRootCmd.AddCommand(
		buildcmds.NewListCommand(clientSetProvider),
		buildcmds.NewStatusCommand(clientSetProvider),
		buildcmds.NewLogsCommand(clientSetProvider),
	)
	return buildRootCmd
}

func getSecretCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	credentialFetcher := &commands.CredentialFetcher{}
	secretFactory := &secret.Factory{
		CredentialFetcher: credentialFetcher,
	}

	secretRootCmd := &cobra.Command{
		Use:     "secret",
		Short:   "Secret Commands",
		Aliases: []string{"secrets"},
	}
	secretRootCmd.AddCommand(
		secretcmds.NewCreateCommand(clientSetProvider, secretFactory),
		secretcmds.NewDeleteCommand(clientSetProvider),
		secretcmds.NewListCommand(clientSetProvider),
	)
	return secretRootCmd
}

func getClusterBuilderCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	clusterBuilderRootCmd := &cobra.Command{
		Use:     "clusterbuilder",
		Short:   "ClusterBuilder Commands",
		Aliases: []string{"clusterbuilders", "clstrbldrs", "clstrbldr", "cbldrs", "cbldr", "cbs", "cb"},
	}
	clusterBuilderRootCmd.AddCommand(
		clusterbuildercmds.NewCreateCommand(clientSetProvider),
		clusterbuildercmds.NewPatchCommand(clientSetProvider),
		clusterbuildercmds.NewSaveCommand(clientSetProvider),
		clusterbuildercmds.NewListCommand(clientSetProvider),
		clusterbuildercmds.NewStatusCommand(clientSetProvider),
		clusterbuildercmds.NewDeleteCommand(clientSetProvider),
	)
	return clusterBuilderRootCmd
}

func getBuilderCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	builderRootCmd := &cobra.Command{
		Use:     "builder",
		Short:   "Builder Commands",
		Aliases: []string{"builders", "bldrs", "bldr"},
	}
	builderRootCmd.AddCommand(
		buildercmds.NewCreateCommand(clientSetProvider),
		buildercmds.NewPatchCommand(clientSetProvider),
		buildercmds.NewSaveCommand(clientSetProvider),
		buildercmds.NewListCommand(clientSetProvider),
		buildercmds.NewDeleteCommand(clientSetProvider),
		buildercmds.NewStatusCommand(clientSetProvider),
	)
	return builderRootCmd
}

func getStackCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	factory := &clusterstack.Factory{}

	stackRootCmd := &cobra.Command{
		Use:     "clusterstack",
		Aliases: []string{"clusterstacks", "clstrcsks", "clstrcsk", "cstks", "cstk", " csks", "csk"},
		Short:   "ClusterStack Commands",
	}
	stackRootCmd.AddCommand(
		clusterstackcmds.NewCreateCommand(clientSetProvider, factory),
		clusterstackcmds.NewUpdateCommand(clientSetProvider, factory),
		clusterstackcmds.NewSaveCommand(clientSetProvider, factory),
		clusterstackcmds.NewListCommand(clientSetProvider),
		clusterstackcmds.NewStatusCommand(clientSetProvider),
		clusterstackcmds.NewDeleteCommand(clientSetProvider),
	)
	stackRootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return configureClusterStackFactory(cmd, factory)
	}

	return stackRootCmd
}

func getStoreCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	factory := &clusterstore.Factory{}

	storeRootCommand := &cobra.Command{
		Use:     "clusterstore",
		Aliases: []string{"clusterstores", "clstrcsrs", "clstrcsr", "csrs", "csr"},
		Short:   "ClusterStore Commands",
	}
	storeRootCommand.AddCommand(
		storecmds.NewCreateCommand(clientSetProvider, factory),
		storecmds.NewAddCommand(clientSetProvider, factory),
		storecmds.NewSaveCommand(clientSetProvider, factory),
		storecmds.NewDeleteCommand(clientSetProvider, commands.NewConfirmationProvider()),
		storecmds.NewStatusCommand(clientSetProvider),
		storecmds.NewRemoveCommand(clientSetProvider),
		storecmds.NewListCommand(clientSetProvider),
	)
	storeRootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return configureClusterStoreFactory(cmd, factory)
	}

	return storeRootCommand
}

func getImportCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	bpUploader := &buildpackage.Uploader{
		Fetcher:   &registry.Fetcher{},
		Relocator: &registry.Relocator{},
	}

	return importcmds.NewImportCommand(
		clientSetProvider,
		bpUploader,
		&registry.Relocator{},
		&registry.Fetcher{},
		commands.Differ{},
		importpkg.DefaultTimestampProvider(),
		commands.NewConfirmationProvider(),
	)
}

func getCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(kp completion bash)

# To load completions for each session, execute once:
Linux:
  $ kp completion bash > /etc/bash_completion.d/kp
MacOS:
  $ kp completion bash > /usr/local/etc/bash_completion.d/kp

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ kp completion zsh > "${fpath[1]}/_kp"

# You will need to start a new shell for this setup to take effect.

Fish:

$ kp completion fish | source

# To load completions for each session, execute once:
$ kp completion fish > ~/.config/fish/completions/kp.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	}
}

func configureImageFactory(cmd *cobra.Command, factory *image.Factory) error {
	dryRun, err := commands.GetBoolFlag(commands.DryRunFlag, cmd)
	if err != nil {
		return err
	}

	if dryRun {
		factory.SourceUploader = registry.DryRunSourceUploader{}
	} else {
		factory.SourceUploader = registry.SourceUploaderImpl{}
	}
	return nil
}

func configureClusterStackFactory(cmd *cobra.Command, factory *clusterstack.Factory) error {
	relocator, err := getImageRelocator(cmd)
	if err != nil {
		return err
	}

	factory.Relocator = relocator
	factory.Fetcher = registry.Fetcher{}
	return nil
}

func configureClusterStoreFactory(cmd *cobra.Command, factory *clusterstore.Factory) error {
	relocator, err := getImageRelocator(cmd)
	if err != nil {
		return err
	}

	factory.Uploader = &buildpackage.Uploader{
		Fetcher:   &registry.Fetcher{},
		Relocator: relocator,
	}
	return nil
}

func getImageRelocator(cmd *cobra.Command) (registry.Relocator, error) {
	var relocator registry.Relocator

	dryRun, err := commands.GetBoolFlag(commands.DryRunFlag, cmd)
	if err != nil {
		return relocator, err
	}

	if dryRun {
		relocator = registry.DryRunRelocator{}
	} else {
		relocator = registry.RelocatorImpl{}
	}
	return relocator, err
}
