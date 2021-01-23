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

	"github.com/pivotal/build-service-cli/pkg/commands"
	buildcmds "github.com/pivotal/build-service-cli/pkg/commands/build"
	buildercmds "github.com/pivotal/build-service-cli/pkg/commands/builder"
	clusterbuildercmds "github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	clusterstackcmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	clusterstorecmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	importcmds "github.com/pivotal/build-service-cli/pkg/commands/import"
	"github.com/pivotal/build-service-cli/pkg/commands/lifecycle"
	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
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
		getLifecycleCommand(clientSetProvider),
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

	imageRootCmd := &cobra.Command{
		Use:     "image",
		Short:   "Image commands",
		Aliases: []string{"images", "imgs", "img"},
	}
	imageRootCmd.AddCommand(
		imgcmds.NewCreateCommand(clientSetProvider, registry.DefaultUtilProvider{}, newImageWaiter),
		imgcmds.NewPatchCommand(clientSetProvider, registry.DefaultUtilProvider{}, newImageWaiter),
		imgcmds.NewSaveCommand(clientSetProvider, registry.DefaultUtilProvider{}, newImageWaiter),
		imgcmds.NewListCommand(clientSetProvider),
		imgcmds.NewDeleteCommand(clientSetProvider),
		imgcmds.NewTriggerCommand(clientSetProvider),
		imgcmds.NewStatusCommand(clientSetProvider),
	)
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
		buildcmds.NewStatusCommand(clientSetProvider, registry.DefaultUtilProvider{}),
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
	stackRootCmd := &cobra.Command{
		Use:     "clusterstack",
		Aliases: []string{"clusterstacks", "clstrcsks", "clstrcsk", "cstacks", "cstack", "cstks", "cstk", "csks", "csk"},
		Short:   "ClusterStack Commands",
	}
	stackRootCmd.AddCommand(
		clusterstackcmds.NewCreateCommand(clientSetProvider, registry.DefaultUtilProvider{}),
		clusterstackcmds.NewUpdateCommand(clientSetProvider, registry.DefaultUtilProvider{}),
		clusterstackcmds.NewSaveCommand(clientSetProvider, registry.DefaultUtilProvider{}),
		clusterstackcmds.NewListCommand(clientSetProvider),
		clusterstackcmds.NewStatusCommand(clientSetProvider),
		clusterstackcmds.NewDeleteCommand(clientSetProvider),
	)
	return stackRootCmd
}

func getStoreCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	storeRootCommand := &cobra.Command{
		Use:     "clusterstore",
		Aliases: []string{"clusterstores", "clstrcsrs", "clstrcsr", "cstores", "cstore", "cstrs", "cstr", "csrs", "csr"},
		Short:   "ClusterStore Commands",
	}
	storeRootCommand.AddCommand(
		clusterstorecmds.NewCreateCommand(clientSetProvider, registry.DefaultUtilProvider{}),
		clusterstorecmds.NewAddCommand(clientSetProvider, registry.DefaultUtilProvider{}),
		clusterstorecmds.NewSaveCommand(clientSetProvider, registry.DefaultUtilProvider{}),
		clusterstorecmds.NewDeleteCommand(clientSetProvider, commands.NewConfirmationProvider()),
		clusterstorecmds.NewStatusCommand(clientSetProvider),
		clusterstorecmds.NewRemoveCommand(clientSetProvider),
		clusterstorecmds.NewListCommand(clientSetProvider),
	)

	return storeRootCommand
}

func getLifecycleCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	lifecycleRootCommand := &cobra.Command{
		Use:     "lifecycle",
		Short:   "Lifecycle Commands",
	}
	lifecycleRootCommand.AddCommand(
		lifecycle.NewUpdateCommand(clientSetProvider, registry.DefaultUtilProvider{}),
	)
	return lifecycleRootCommand
}

func getImportCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {

	return importcmds.NewImportCommand(
		commands.Differ{},
		clientSetProvider,
		registry.DefaultUtilProvider{},
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
