// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"bytes"
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
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

	fakeImageWaiter := &fakes.FakeImageWaiter{}
	waiterFunc := func(set k8s.ClientSet) image.ImageWaiter {
		return fakeImageWaiter
	}

	when("a namespace is provided", func() {
		when("an image build is available", func() {
			it("triggers the latest build", func() {
				clientSet := fake.NewSimpleClientset(testNamespacedBuilds...)
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider, waiterFunc)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", namespace})

				err := cmd.Execute()
				require.NoError(t, err)
				require.Equal(t, "\"some-image\" triggered\n", out.String())

				actions, err := testhelpers.ActionRecorderList{clientSet}.ActionsByVerb()
				require.NoError(t, err)

				require.Len(t, actions.Updates, 1)
				build := actions.Updates[0].GetObject().(*v1alpha1.Build)
				require.Equal(t, build.Name, "build-three")
				require.NotEmpty(t, build.Annotations[image.BuildNeededAnnotation])
				require.Len(t, fakeImageWaiter.Calls, 0)
			})

			it("waits when the wait flag is used", func() {
				img := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: namespace,
					},
				}

				clientSet := fake.NewSimpleClientset(append(testNamespacedBuilds, img)...)
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider, waiterFunc)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", namespace, "--wait"})

				err := cmd.Execute()
				require.NoError(t, err)
				require.Equal(t, "\"some-image\" triggered\n", out.String())

				actions, err := testhelpers.ActionRecorderList{clientSet}.ActionsByVerb()
				require.NoError(t, err)

				require.Len(t, actions.Updates, 1)
				build := actions.Updates[0].GetObject().(*v1alpha1.Build)
				require.Equal(t, build.Name, "build-three")
				require.NotEmpty(t, build.Annotations[image.BuildNeededAnnotation])
				require.Len(t, fakeImageWaiter.Calls, 1)
				require.Equal(t, fakeImageWaiter.Calls[0], img)
			})
		})

		when("an image build is not available", func() {
			it("returns an error", func() {
				clientSet := fake.NewSimpleClientset()
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider, waiterFunc)

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
				clientSet := fake.NewSimpleClientset(testBuilds...)
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider, waiterFunc)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", defaultNamespace})

				err := cmd.Execute()
				require.NoError(t, err)
				require.Equal(t, "\"some-image\" triggered\n", out.String())

				actions, err := testhelpers.ActionRecorderList{clientSet}.ActionsByVerb()
				require.NoError(t, err)

				require.Len(t, actions.Updates, 1)
				build := actions.Updates[0].GetObject().(*v1alpha1.Build)
				require.Equal(t, build.Name, "build-three")
				require.NotEmpty(t, build.Annotations[image.BuildNeededAnnotation])
				require.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		when("an image build is not available", func() {
			it("returns an error", func() {
				clientSet := fake.NewSimpleClientset()
				clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
				cmd := image.NewTriggerCommand(clientSetProvider, waiterFunc)

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				cmd.SetArgs([]string{"some-image", "-n", defaultNamespace})

				err := cmd.Execute()
				require.EqualError(t, err, "no builds found")
			})
		})
	})
}
