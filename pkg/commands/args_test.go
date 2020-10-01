package commands_test

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

func TestExactArgsWithUsage(t *testing.T) {
	cmd := cobra.Command{
		Args: commands.ExactArgsWithUsage(1),
	}
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "some usage")
		return err
	})

	expectedMsg := "accepts 1 arg(s), received 0\n\nsome usage\n"
	err := cmd.ValidateArgs([]string{})

	if err == nil || err.Error() != expectedMsg {
		t.Errorf(`Did not return expected usage error from using wrong number of args. 
Expected:
%v
Actual: 
%v`, expectedMsg, err)
	}
}
