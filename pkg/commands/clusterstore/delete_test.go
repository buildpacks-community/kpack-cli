// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"fmt"
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStoreDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreDeleteCommand", testClusterStoreDeleteCommand)
}

func testClusterStoreDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		storeName = "some-store-name"
	)

	var confirmationProvider FakeConfirmationProvider

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterstore.NewDeleteCommand(clientSetProvider, &confirmationProvider)
	}

	when("confirmation is given by user", func() {
		it.Before(func() {
			confirmationProvider.confirm = true
			confirmationProvider.err = nil
		})

		when("store exists", func() {
			store := &expv1alpha1.ClusterStore{
				ObjectMeta: v1.ObjectMeta{
					Name: storeName,
					Annotations: map[string]string{
						"buildservice.pivotal.io/defaultRepository": "some/path",
					},
				},
				Spec: expv1alpha1.ClusterStoreSpec{
					Sources: []expv1alpha1.StoreImage{
						{
							Image: "some/imageInStore",
						},
					},
				},
			}

			it.Before(func() {
				confirmationProvider.requested = false
			})

			it("confirms and deletes the store", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						store,
					},
					Args:           []string{storeName},
					ExpectErr:      false,
					ExpectedOutput: fmt.Sprintf("%q store deleted\n", storeName),
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							Name: storeName,
						},
					},
				}.TestKpack(t, cmdFunc)
				assert.True(t, confirmationProvider.requested)
			})
		})

		when("store does not exist", func() {
			it.Before(func() {
				confirmationProvider.requested = false
			})

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
				assert.True(t, confirmationProvider.requested)
			})
		})
	})

	when("confirmation is not given by user", func() {
		it.Before(func() {
			confirmationProvider.confirm = false
			confirmationProvider.err = nil
			confirmationProvider.requested = false
		})

		it("skips deleting the store", func() {
			testhelpers.CommandTest{
				Objects:        nil,
				Args:           []string{storeName},
				ExpectErr:      false,
				ExpectedOutput: "Skipping store deletion\n",
			}.TestKpack(t, cmdFunc)
			assert.True(t, confirmationProvider.requested)
		})
	})

	when("confirmation process errors", func() {
		confirmationError := errors.New("some weird error")
		it.Before(func() {
			confirmationProvider.confirm = false
			confirmationProvider.err = confirmationError
			confirmationProvider.requested = false
		})

		it("confirms and bubbles up the error", func() {
			testhelpers.CommandTest{
				Objects:        nil,
				Args:           []string{storeName},
				ExpectErr:      true,
				ExpectedOutput: fmt.Sprintf("Error: %s\n", confirmationError),
			}.TestKpack(t, cmdFunc)
			assert.True(t, confirmationProvider.requested)
		})
	})

	when("force deletion flag is used", func() {
		when("store exists", func() {
			store := &expv1alpha1.ClusterStore{
				ObjectMeta: v1.ObjectMeta{
					Name: storeName,
					Annotations: map[string]string{
						"buildservice.pivotal.io/defaultRepository": "some/path",
					},
				},
				Spec: expv1alpha1.ClusterStoreSpec{
					Sources: []expv1alpha1.StoreImage{
						{
							Image: "some/imageInStore",
						},
					},
				},
			}

			it.Before(func() {
				confirmationProvider.requested = false
			})

			it("deletes the store without confirmation", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						store,
					},
					Args:           []string{storeName, "-f"},
					ExpectErr:      false,
					ExpectedOutput: fmt.Sprintf("%q store deleted\n", storeName),
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							Name: storeName,
						},
					},
				}.TestKpack(t, cmdFunc)
				assert.False(t, confirmationProvider.requested)
			})
		})

		when("store does not exist", func() {
			it.Before(func() {
				confirmationProvider.requested = false
			})

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
				assert.False(t, confirmationProvider.requested)
			})
		})
	})
}

type FakeConfirmationProvider struct {
	// return values for confirm request
	confirm bool
	err     error
	// tracks if confirmation was requested
	requested bool
}

func (f *FakeConfirmationProvider) Confirm(_ string, _ ...string) (bool, error) {
	f.requested = true
	return f.confirm, f.err
}
