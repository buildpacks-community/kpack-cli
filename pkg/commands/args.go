package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ExactArgsWithUsage(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d\n\n%s", n, len(args), cmd.UsageString())
		}
		return nil
	}
}
func OptionalArgsWithUsage(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 && len(args) != n {
			return fmt.Errorf("accepts 0 or %d arg(s), received %d\n\n%s", n, len(args), cmd.UsageString())
		}
		return nil
	}
}
