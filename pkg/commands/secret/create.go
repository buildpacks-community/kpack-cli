package secret

import (
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

func NewCreateCommand(kpackClient versioned.Interface, defaultNamespace string) *cobra.Command {
	return nil
}
