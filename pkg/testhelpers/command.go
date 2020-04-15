package testhelpers

import (
	"bytes"
	"testing"

	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
)

type CommandTest struct {
	Objects []runtime.Object

	StdIn string
	Args  []string

	ExpectErr      bool
	ExpectedOutput string
	ExpectUpdates  []clientgotesting.UpdateActionImpl
	ExpectCreates  []runtime.Object
	ExpectDeletes  []string
}

func (c CommandTest) TestKpack(t *testing.T, cmdFactory func(clientSet *kpackfakes.Clientset) *cobra.Command) {
	t.Helper()
	client := kpackfakes.NewSimpleClientset(c.Objects...)

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
	TestKpackActions(t, client, c.ExpectUpdates, c.ExpectCreates, c.ExpectDeletes)
}

func (c CommandTest) TestK8s(t *testing.T, cmdFactory func(clientSet *k8sfakes.Clientset) *cobra.Command) {
	t.Helper()
	client := k8sfakes.NewSimpleClientset(c.Objects...)

	cmd := cmdFactory(client)
	cmd.SetArgs(c.Args)

	inputBuffer := bytes.NewBufferString(c.StdIn)
	cmd.SetIn(inputBuffer)

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	err := cmd.Execute()
	if !c.ExpectErr {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}

	require.Equal(t, c.ExpectedOutput, out.String())
	TestK8sActions(t, client, c.ExpectUpdates, c.ExpectCreates, c.ExpectDeletes)
}
