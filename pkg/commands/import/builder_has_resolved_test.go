// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/watch"
)

func TestBuilderHasResolved(t *testing.T) {
	spec.Run(t, "TestBuilderHasResolved", testBuilderHasResolved)
}

func testBuilderHasResolved(t *testing.T, when spec.G, it spec.S) {
	var (
		cb              *v1alpha1.ClusterBuilder
		storeGeneration int64 = 2
		stackGeneration int64 = 2
	)

	it.Before(func() {
		cb = &v1alpha1.ClusterBuilder{}
	})

	it("returns true when observed store and stack gen are up to date", func() {
		cb.Status = v1alpha1.BuilderStatus{
			ObservedStoreGeneration: storeGeneration,
			ObservedStackGeneration: stackGeneration,
		}

		e := watch.Event{
			Object: cb,
		}
		done, err := builderHasResolved(storeGeneration, stackGeneration)(e)
		require.NoError(t, err)
		require.True(t, done)
	})

	it("returns true when observed store and stack gen are 0", func() {
		cb.Status = v1alpha1.BuilderStatus{
			ObservedStoreGeneration: 0,
			ObservedStackGeneration: 0,
		}

		e := watch.Event{
			Object: cb,
		}
		done, err := builderHasResolved(storeGeneration, stackGeneration)(e)
		require.NoError(t, err)
		require.True(t, done)
	})

	it("returns false when observed store gen is not up to date", func() {
		cb.Status = v1alpha1.BuilderStatus{
			ObservedStoreGeneration: storeGeneration - 1,
			ObservedStackGeneration: stackGeneration,
		}

		e := watch.Event{
			Object: cb,
		}
		done, err := builderHasResolved(storeGeneration, stackGeneration)(e)
		require.NoError(t, err)
		require.False(t, done)
	})

	it("returns false when observed stack gen is not up to date", func() {
		cb.Status = v1alpha1.BuilderStatus{
			ObservedStoreGeneration: storeGeneration,
			ObservedStackGeneration: stackGeneration - 1,
		}

		e := watch.Event{
			Object: cb,
		}
		done, err := builderHasResolved(storeGeneration, stackGeneration)(e)
		require.NoError(t, err)
		require.False(t, done)
	})
}
