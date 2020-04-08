package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/image"
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

	imageLister := &image.Lister{
		KpackClient: kpackClient,
	}

	imageDeleter := &image.Deleter{
		KpackClient: kpackClient,
	}

	imageRootCmd := &cobra.Command{
		Use:   "image",
		Short: "Image commands",
	}
	imageRootCmd.AddCommand(
		imgcmds.NewGetCommand(kpackClient, defaultNamespace),
		imgcmds.NewApplyCommand(kpackClient, defaultNamespace),
		imgcmds.NewListCommand(os.Stdout, defaultNamespace, imageLister),
		imgcmds.NewDeleteCommand(os.Stdout, defaultNamespace, imageDeleter),
	)

	rootCmd := &cobra.Command{
		Use: "tbctl",
	}
	rootCmd.AddCommand(
		imageRootCmd,
	)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
