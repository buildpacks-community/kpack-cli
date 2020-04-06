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

	imageApplier := &image.Applier{
		DefaultNamespace: defaultNamespace,
		KpackClient:      kpackClient,
	}

	imageRootCmd := &cobra.Command{
		Use:   "image",
		Short: "Image commands",
	}
	imageRootCmd.AddCommand(
		imgcmds.NewApplyCommand(os.Stdout, imageApplier),
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
