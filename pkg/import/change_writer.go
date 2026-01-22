// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/config"
	buildk8s "github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

type changeWriter interface {
	writeDiff(diffs string) error
	writeChange(header string)
}

func writeClusterLifecyclesChange(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, lifecycles []ClusterLifecycle, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, lifecycle := range lifecycles {
		oldLifecycle, err := cs.KpackClient.KpackV1alpha2().ClusterLifecycles().Get(ctx, lifecycle.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldLifecycle = nil
		}

		diff, err := differ.DiffClusterLifecycle(keychain, kpConfig, oldLifecycle, lifecycle)
		if err != nil {
			return err
		}
		if err = cw.writeDiff(diff); err != nil {
			return err
		}
	}

	cw.writeChange("ClusterLifecycles")
	return nil
}

func writeClusterBuildpacksChange(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, buildpacks []ClusterBuildpack, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, buildpack := range buildpacks {
		oldBuildpack, err := cs.KpackClient.KpackV1alpha2().ClusterBuildpacks().Get(ctx, buildpack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldBuildpack = nil
		}

		diff, err := differ.DiffClusterBuildpack(keychain, kpConfig, oldBuildpack, buildpack)
		if err != nil {
			return err
		}
		if err = cw.writeDiff(diff); err != nil {
			return err
		}
	}

	cw.writeChange("ClusterBuildpacks")
	return nil
}

func writeClusterStoresChange(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, stores []ClusterStore, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, store := range stores {
		oldStore, err := cs.KpackClient.KpackV1alpha2().ClusterStores().Get(ctx, store.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldStore = nil
		}

		diff, err := differ.DiffClusterStore(keychain, kpConfig, oldStore, store)
		if err != nil {
			return err
		}
		if err = cw.writeDiff(diff); err != nil {
			return err
		}
	}

	cw.writeChange("ClusterStores")
	return nil
}

func writeClusterStacksChange(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, stacks []ClusterStack, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, stack := range stacks {
		oldStack, err := cs.KpackClient.KpackV1alpha2().ClusterStacks().Get(ctx, stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldStack = nil
		}

		diff, err := differ.DiffClusterStack(keychain, kpConfig, oldStack, stack)
		if err != nil {
			return err
		}
		if err = cw.writeDiff(diff); err != nil {
			return err
		}
	}

	cw.writeChange("ClusterStacks")
	return nil
}

func writeClusterBuildersChange(ctx context.Context, builders []ClusterBuilder, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, builder := range builders {
		oldBuilder, err := cs.KpackClient.KpackV1alpha2().ClusterBuilders().Get(ctx, builder.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldBuilder = nil
		}

		diff, err := differ.DiffClusterBuilder(oldBuilder, builder)
		if err != nil {
			return err
		}
		if err = cw.writeDiff(diff); err != nil {
			return err
		}
	}

	cw.writeChange("ClusterBuilders")
	return nil
}
