// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestImageDeleteCommand(t *testing.T) {
	spec.Run(t, "TestImageDeleteCommand", testImageDeleteCommand)
}

func testImageDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return image.NewDeleteCommand(clientSetProvider)
	}

	when("a namespace is provided", func() {
		when("an image is available", func() {
			it("deletes the image", func() {
				image := &v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-image",
						Namespace: "some-namespace",
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image,
					},
					Args: []string{"some-image", "-n", "some-namespace"},
					ExpectedOutput: `Image "some-image" deleted
`,
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "some-namespace",
							},
							Name: image.Name,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})

		when("an image is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"some-image", "-n", "some-namespace"},
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "some-namespace",
							},
							Name: "some-image",
						},
					},
					ExpectedOutput: "Error: images.kpack.io \"some-image\" not found\n",
					ExpectErr:      true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("a namespace is not provided", func() {
		when("an image is available", func() {
			it("deletes the image", func() {
				image := &v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image,
					},
					Args: []string{"some-image"},
					ExpectedOutput: `Image "some-image" deleted
`,
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: defaultNamespace,
							},
							Name: image.Name,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})

		when("an image is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"some-image", "-n", "some-namespace"},
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "some-namespace",
							},
							Name: "some-image",
						},
					},
					ExpectedOutput: "Error: images.kpack.io \"some-image\" not found\n",
					ExpectErr:      true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}
