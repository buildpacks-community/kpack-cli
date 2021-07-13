// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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
	Objects      []runtime.Object
	K8sObjects   []runtime.Object
	KpackObjects []runtime.Object

	StdIn string
	Args  []string

	ExpectErr           bool
	ExpectedOutput      string
	ExpectedErrorOutput string
	ExpectUpdates       []clientgotesting.UpdateActionImpl
	ExpectCreates       []runtime.Object
	ExpectDeletes       []clientgotesting.DeleteActionImpl
	ExpectPatches       []string
}

func (c CommandTest) TestK8sAndKpack(t *testing.T, cmdFactory func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command) {
	t.Helper()
	k8sClient := k8sfakes.NewSimpleClientset(c.K8sObjects...)
	kpackClient := kpackfakes.NewSimpleClientset(c.KpackObjects...)

	cmd := cmdFactory(k8sClient, kpackClient)
	cmd.SetArgs(c.Args)

	inputBuffer := bytes.NewBufferString(c.StdIn)
	cmd.SetIn(inputBuffer)

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if !c.ExpectErr {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}

	require.Equal(t, c.ExpectedOutput, out.String(), "Actual output does not match ExpectedOutput")
	require.Equal(t, c.ExpectedErrorOutput, errOut.String(), "Actual error output does not match ExpectedErrorOutput")
	TestK8sAndKpackActions(t, k8sClient, kpackClient, c.ExpectUpdates, c.ExpectCreates, c.ExpectDeletes, c.ExpectPatches)
}

func (c CommandTest) TestKpack(t *testing.T, cmdFactory func(clientSet *kpackfakes.Clientset) *cobra.Command) {
	t.Helper()
	client := kpackfakes.NewSimpleClientset(c.Objects...)

	cmd := cmdFactory(client)
	cmd.SetArgs(c.Args)

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if !c.ExpectErr {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}

	require.Equal(t, c.ExpectedOutput, out.String(), "Actual output does not match ExpectedOutput")
	require.Equal(t, c.ExpectedErrorOutput, errOut.String(), "Actual error output does not match ExpectedErrorOutput")
	TestKpackActions(t, client, c.ExpectUpdates, c.ExpectCreates, c.ExpectDeletes, c.ExpectPatches)
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

	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if !c.ExpectErr {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}

	require.Equal(t, c.ExpectedOutput, out.String(), "Actual output does not match ExpectedOutput")
	require.Equal(t, c.ExpectedErrorOutput, errOut.String(), "Actual error output does not match ExpectedErrorOutput")
	TestK8sActions(t, client, c.ExpectUpdates, c.ExpectCreates, c.ExpectDeletes, c.ExpectPatches)
}
