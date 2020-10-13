// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"fmt"
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStoreDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreDeleteCommand", testClusterStoreDeleteCommand)
}

func testClusterStoreDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		storeName = "some-store-name"
	)

	var confirmationProvider *fakes.FakeConfirmationProvider

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterstore.NewDeleteCommand(clientSetProvider, confirmationProvider)
	}

	it.Before(func() {
		confirmationProvider = fakes.NewFakeConfirmationProvider(true, nil)
	})

	when("confirmation is given by user", func() {
		when("store exists", func() {
			store := &v1alpha1.ClusterStore{
				ObjectMeta: v1.ObjectMeta{
					Name: storeName,
				},
				Spec: v1alpha1.ClusterStoreSpec{
					Sources: []v1alpha1.StoreImage{
						{
							Image: "some/imageInStore",
						},
					},
				},
			}

			it("confirms and deletes the store", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						store,
					},
					Args:      []string{storeName},
					ExpectErr: false,
					ExpectedOutput: `ClusterStore "some-store-name" store deleted
`,
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							Name: storeName,
						},
					},
				}.TestKpack(t, cmdFunc)
				assert.True(t, confirmationProvider.WasRequested())
			})
		})

		when("store does not exist", func() {
			it("confirms and errors with store not found", func() {
				testhelpers.CommandTest{
					Objects:        nil,
					Args:           []string{storeName},
					ExpectErr:      true,
					ExpectedOutput: fmt.Sprintf("Error: Store %q does not exist\n", storeName),
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							Name: storeName,
						},
					},
				}.TestKpack(t, cmdFunc)
				assert.True(t, confirmationProvider.WasRequested())
			})
		})
	})

	when("confirmation is not given by user", func() {
		it.Before(func() {
			confirmationProvider = fakes.NewFakeConfirmationProvider(false, nil)
		})

		it("skips deleting the store", func() {
			testhelpers.CommandTest{
				Objects:        nil,
				Args:           []string{storeName},
				ExpectErr:      false,
				ExpectedOutput: "Skipping ClusterStore deletion\n",
			}.TestKpack(t, cmdFunc)
			assert.True(t, confirmationProvider.WasRequested())
		})
	})

	when("confirmation process errors", func() {
		confirmationError := errors.New("some weird error")
		it.Before(func() {
			confirmationProvider = fakes.NewFakeConfirmationProvider(false, confirmationError)
		})

		it("confirms and bubbles up the error", func() {
			testhelpers.CommandTest{
				Objects:        nil,
				Args:           []string{storeName},
				ExpectErr:      true,
				ExpectedOutput: fmt.Sprintf("Error: %s\n", confirmationError),
			}.TestKpack(t, cmdFunc)
			assert.True(t, confirmationProvider.WasRequested())
		})
	})

	when("force deletion flag is used", func() {
		when("store exists", func() {
			store := &v1alpha1.ClusterStore{
				ObjectMeta: v1.ObjectMeta{
					Name: storeName,
				},
				Spec: v1alpha1.ClusterStoreSpec{
					Sources: []v1alpha1.StoreImage{
						{
							Image: "some/imageInStore",
						},
					},
				},
			}

			it("deletes the store without confirmation", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						store,
					},
					Args:      []string{storeName, "-f"},
					ExpectErr: false,
					ExpectedOutput: `ClusterStore "some-store-name" store deleted
`,
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							Name: storeName,
						},
					},
				}.TestKpack(t, cmdFunc)
				assert.False(t, confirmationProvider.WasRequested())
			})
		})

		when("store does not exist", func() {
			it("does not confirm and errors with store not found", func() {
				testhelpers.CommandTest{
					Objects:        nil,
					Args:           []string{storeName, "--force"},
					ExpectErr:      true,
					ExpectedOutput: fmt.Sprintf("Error: Store %q does not exist\n", storeName),
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							Name: storeName,
						},
					},
				}.TestKpack(t, cmdFunc)
				assert.False(t, confirmationProvider.WasRequested())
			})
		})
	})
}
