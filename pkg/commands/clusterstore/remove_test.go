// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterstore"
	commandsfakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

func TestClusterStoreRemoveCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreRemoveCommand", testClusterStoreRemoveCommand)
}

func testClusterStoreRemoveCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		storeName     = "some-store"
		image1InStore = "some/imageinStore1@sha256:1231alreadyInStore"
		image2InStore = "some/imageinStore2@sha256:1232alreadyInStore"
	)

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterstore.NewRemoveCommand(clientSetProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	store := &v1alpha2.ClusterStore{
		ObjectMeta: v1.ObjectMeta{
			Name: storeName,
		},
		Spec: v1alpha2.ClusterStoreSpec{
			Sources: []corev1alpha1.ImageSource{
				{
					Image: image1InStore,
				},
				{
					Image: image2InStore,
				},
			},
		},
		Status: v1alpha2.ClusterStoreStatus{
			Buildpacks: []corev1alpha1.BuildpackStatus{
				{
					BuildpackInfo: corev1alpha1.BuildpackInfo{
						Id:      "some-buildpackage",
						Version: "1.2.3",
					},
					StoreImage: corev1alpha1.ImageSource{
						Image: image1InStore,
					},
				},
				{
					BuildpackInfo: corev1alpha1.BuildpackInfo{
						Id:      "another-buildpackage",
						Version: "4.5.6",
					},
					StoreImage: corev1alpha1.ImageSource{
						Image: image2InStore,
					},
				},
			},
		},
	}

	it("removes a single buildpackage from the store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"--buildpackage", "some-buildpackage@1.2.3",
			},
			ExpectPatches: []string{
				`{"spec":{"sources":[{"image":"some/imageinStore2@sha256:1232alreadyInStore"}]}}`,
			},
			ExpectedOutput: `Removing Buildpackages...
Removing buildpackage some-buildpackage@1.2.3
ClusterStore "some-store" updated
`,
		}.TestKpack(t, cmdFunc)
		require.Len(t, fakeWaiter.WaitCalls, 1)
	})

	it("removes multiple buildpackages from the store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "some-buildpackage@1.2.3",
				"-b", "another-buildpackage@4.5.6",
			},
			ExpectPatches: []string{
				`{"spec":{"sources":null}}`,
			},
			ExpectedOutput: `Removing Buildpackages...
Removing buildpackage some-buildpackage@1.2.3
Removing buildpackage another-buildpackage@4.5.6
ClusterStore "some-store" updated
`,
		}.TestKpack(t, cmdFunc)
	})

	it("fails if the provided store does not exist", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				"invalid-store",
				"-b", "some-buildpackage@1.2.3",
				"-b", "another-buildpackage@4.5.6",
			},
			ExpectErr:           true,
			ExpectedErrorOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
		}.TestKpack(t, cmdFunc)
	})

	it("fails if even one buildpackage is not in the store but the rest are", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "some-buildpackage@1.2.3",
				"-b", "does-not-exist-buildpackage@7.8.9",
			},
			ExpectErr:           true,
			ExpectedOutput:      "Removing Buildpackages...\n",
			ExpectedErrorOutput: "Error: Buildpackage 'does-not-exist-buildpackage@7.8.9' does not exist in the ClusterStore\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error if buildpackage does not exist in store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "does-not-exist-buildpackage@7.8.9",
			},
			ExpectErr:           true,
			ExpectedOutput:      "Removing Buildpackages...\n",
			ExpectedErrorOutput: "Error: Buildpackage 'does-not-exist-buildpackage@7.8.9' does not exist in the ClusterStore\n",
		}.TestKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: some-store
spec:
  sources:
  - image: some/imageinStore2@sha256:1232alreadyInStore
status:
  buildpacks:
  - buildpackage: {}
    id: some-buildpackage
    storeImage:
      image: some/imageinStore1@sha256:1231alreadyInStore
    version: 1.2.3
  - buildpackage: {}
    id: another-buildpackage
    storeImage:
      image: some/imageinStore2@sha256:1232alreadyInStore
    version: 4.5.6
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some-buildpackage@1.2.3",
					"--output", "yaml",
				},
				ExpectPatches: []string{
					`{"spec":{"sources":[{"image":"some/imageinStore2@sha256:1232alreadyInStore"}]}}`,
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Removing Buildpackages...
Removing buildpackage some-buildpackage@1.2.3
`,
			}.TestKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = "{\n    \"kind\": \"ClusterStore\",\n    \"apiVersion\": \"kpack.io/v1alpha2\",\n    \"metadata\": {\n        \"name\": \"some-store\",\n        \"creationTimestamp\": null\n    },\n    \"spec\": {\n        \"sources\": [\n            {\n                \"image\": \"some/imageinStore2@sha256:1232alreadyInStore\"\n            }\n        ]\n    },\n    \"status\": {\n        \"buildpacks\": [\n            {\n                \"id\": \"some-buildpackage\",\n                \"version\": \"1.2.3\",\n                \"buildpackage\": {},\n                \"storeImage\": {\n                    \"image\": \"some/imageinStore1@sha256:1231alreadyInStore\"\n                }\n            },\n            {\n                \"id\": \"another-buildpackage\",\n                \"version\": \"4.5.6\",\n                \"buildpackage\": {},\n                \"storeImage\": {\n                    \"image\": \"some/imageinStore2@sha256:1232alreadyInStore\"\n                }\n            }\n        ]\n    }\n}\n"

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some-buildpackage@1.2.3",
					"--output", "json",
				},
				ExpectPatches: []string{
					`{"spec":{"sources":[{"image":"some/imageinStore2@sha256:1232alreadyInStore"}]}}`,
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Removing Buildpackages...
Removing buildpackage some-buildpackage@1.2.3
`,
			}.TestKpack(t, cmdFunc)
		})

		it("errors when the provided store does not exist", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					"invalid-store",
					"-b", "some-buildpackage@1.2.3",
					"-b", "another-buildpackage@4.5.6",
					"--output", "yaml",
				},
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
			}.TestKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not remove a buildpackage and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some-buildpackage@1.2.3",
					"--dry-run",
				},
				ExpectedOutput: `Removing Buildpackages... (dry run)
Removing buildpackage some-buildpackage@1.2.3
ClusterStore "some-store" updated (dry run)
`,
			}.TestKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 0)
		})

		it("errors when the provided store does not exist", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					"invalid-store",
					"-b", "some-buildpackage@1.2.3",
					"-b", "another-buildpackage@4.5.6",
					"--dry-run",
				},
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update a clusterstore and prints the resource output in requested format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: some-store
spec:
  sources:
  - image: some/imageinStore2@sha256:1232alreadyInStore
status:
  buildpacks:
  - buildpackage: {}
    id: some-buildpackage
    storeImage:
      image: some/imageinStore1@sha256:1231alreadyInStore
    version: 1.2.3
  - buildpackage: {}
    id: another-buildpackage
    storeImage:
      image: some/imageinStore2@sha256:1232alreadyInStore
    version: 4.5.6
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						store,
					},
					Args: []string{
						storeName,
						"--buildpackage", "some-buildpackage@1.2.3",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Removing Buildpackages... (dry run)
Removing buildpackage some-buildpackage@1.2.3
`,
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}
