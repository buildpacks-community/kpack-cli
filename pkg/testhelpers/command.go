package testhelpers

import (
	"bytes"
	"testing"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

type CommandTest struct {
	Objects []runtime.Object

	Args []string

	ExpectErr      bool
	ExpectedOutput string
	ExpectUpdates  []clientgotesting.UpdateActionImpl
	ExpectCreates  []runtime.Object
}

func (c CommandTest) Test(t *testing.T, cmdFactory func(clientSet *fake.Clientset) *cobra.Command) {
	t.Helper()
	client := fake.NewSimpleClientset(c.Objects...)

	cmd := cmdFactory(client)
	cmd.SetArgs(c.Args)

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	err := cmd.Execute()
	if !c.ExpectErr {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}

	require.Equal(t, c.ExpectedOutput, out.String())
	TestUpdatesAndCreates(t, client, c.ExpectUpdates, c.ExpectCreates)
}
