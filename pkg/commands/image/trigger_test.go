// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestImageTrigger(t *testing.T) {
	spec.Run(t, "TestImageTrigger", testImageTrigger)
}

func testImageTrigger(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
		namespace        = "some-namespace"
	)

	testBuilds := testhelpers.MakeTestBuilds("some-image", defaultNamespace)
	testNamespacedBuilds := testhelpers.MakeTestBuilds("some-image", namespace)

	when("a namespace is provided", func() {
		when("an image build is available", func() {
			it("triggers the latest build", func() {
				clientSet := fake.NewSimpleClientset(testhelpers.BuildsToRuntimeObjs(testNamespacedBuilds)...)
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", namespace})

				err := cmd.Execute()
				require.NoError(t, err)
				require.Equal(t, "Triggered build for Image Resource \"some-image\" with Build Number 3\n", out.String())

				actions, err := testhelpers.ActionRecorderList{clientSet}.ActionsByVerb()
				require.NoError(t, err)

				require.Len(t, actions.Patches, 1)

				var patch buildNeededPatch
				err = json.Unmarshal(actions.Patches[0].GetPatch(), &patch)
				require.NoError(t, err)

				require.Equal(t, actions.Patches[0].GetName(), "build-three")
				require.NotEmpty(t, patch.Metadata.Annotations.BuildNeededAnnotation)
			})
		})

		when("an image build is not available", func() {
			it("returns an error", func() {
				clientSet := fake.NewSimpleClientset()
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", namespace})

				err := cmd.Execute()
				require.EqualError(t, err, "no builds found")
			})
		})
	})

	when("a namespace is not provided", func() {
		when("an image build is available", func() {
			it("triggers the latest build", func() {
				clientSet := fake.NewSimpleClientset(testhelpers.BuildsToRuntimeObjs(testBuilds)...)
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", defaultNamespace})

				err := cmd.Execute()
				require.NoError(t, err)
				require.Equal(t, "Triggered build for Image Resource \"some-image\" with Build Number 3\n", out.String())

				actions, err := testhelpers.ActionRecorderList{clientSet}.ActionsByVerb()
				require.NoError(t, err)

				require.Len(t, actions.Patches, 1)

				var patch buildNeededPatch
				err = json.Unmarshal(actions.Patches[0].GetPatch(), &patch)
				require.NoError(t, err)

				require.Equal(t, actions.Patches[0].GetName(), "build-three")
				require.NotEmpty(t, patch.Metadata.Annotations.BuildNeededAnnotation)
			})
		})

		when("an image build is not available", func() {
			it("returns an error", func() {
				clientSet := fake.NewSimpleClientset()
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", defaultNamespace})

				err := cmd.Execute()
				require.EqualError(t, err, "no builds found")
			})
		})
	})
}

type buildNeededPatch struct {
	Metadata struct {
		Annotations struct {
			BuildNeededAnnotation string `json:"image.kpack.io/additionalBuildNeeded"`
		} `json:"annotations"`
	} `json:"metadata"`
}
