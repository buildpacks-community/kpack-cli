// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"bytes"
	"testing"

	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	kpacktesthelpers "github.com/pivotal/kpack/pkg/reconciler/testhelpers"
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
	listers := kpacktesthelpers.NewListers(c.Objects)

	k8sClient := k8sfakes.NewSimpleClientset(listers.GetKubeObjects()...)
	kpackClient := kpackfakes.NewSimpleClientset(listers.BuildServiceObjects()...)

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
	c.TestK8sAndKpack(t, func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		return cmdFactory(kpackClientSet)
	})
}

func (c CommandTest) TestK8s(t *testing.T, cmdFactory func(clientSet *k8sfakes.Clientset) *cobra.Command) {
	t.Helper()
	c.TestK8sAndKpack(t, func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		return cmdFactory(k8sClientSet)
	})
}
