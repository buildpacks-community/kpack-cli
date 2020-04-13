package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/k8s"
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

	imageRootCmd := &cobra.Command{
		Use:   "image",
		Short: "Image commands",
	}
	imageRootCmd.AddCommand(
		imgcmds.NewGetCommand(kpackClient, defaultNamespace),
		imgcmds.NewApplyCommand(kpackClient, defaultNamespace),
		imgcmds.NewListCommand(kpackClient, defaultNamespace),
		imgcmds.NewDeleteCommand(kpackClient, defaultNamespace),
	)

	secretRootCmd := &cobra.Command{
		Use:   "secret",
		Short: "Secret Commands",
	}
	secretRootCmd.AddCommand(
		secretcmds.NewCreateCommand(k8sClient, defaultNamespace),
	)

	rootCmd := &cobra.Command{
		Use: "tbctl",
	}
	rootCmd.AddCommand(
		imageRootCmd,
		secretRootCmd,
	)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
