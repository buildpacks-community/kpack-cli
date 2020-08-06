// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/builder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuilderDeleteCommand(t *testing.T) {
	spec.Run(t, "TestBuilderDeleteCommand", testBuilderDeleteCommand)
}

func testBuilderDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return builder.NewDeleteCommand(clientSetProvider)
	}

	when("a namespace has been provided", func() {
		when("a builder is available", func() {
			it("deletes the builder", func() {
				builder := &v1alpha1.Builder{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-builder",
						Namespace: "test-namespace",
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						builder,
					},
					Args:           []string{"-n", "test-namespace", "some-builder"},
					ExpectedOutput: "\"some-builder\" deleted\n",
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "test-namespace",
							},
							Name: builder.Name,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})
		when("a builder is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"-n", "test-namespace", "some-builder"},
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "test-namespace",
							},

							Name: "some-builder",
						},
					},
					ExpectedOutput: "Error: builders.kpack.io \"some-builder\" not found\n",
					ExpectErr:      true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("a namespace has not been provided", func() {
		when("a builder is available", func() {
			it("deletes the builder", func() {
				builder := &v1alpha1.Builder{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-builder",
						Namespace: defaultNamespace,
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						builder,
					},
					Args:           []string{"some-builder"},
					ExpectedOutput: "\"some-builder\" deleted\n",
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: defaultNamespace,
							},

							Name: builder.Name,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})
		when("a builder is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"some-builder"},
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: defaultNamespace,
							},

							Name: "some-builder",
						},
					},
					ExpectedOutput: "Error: builders.kpack.io \"some-builder\" not found\n",
					ExpectErr:      true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

}
