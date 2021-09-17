// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"

	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	buildk8s "github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func SummarizeChange(
	ctx context.Context,
	keychain authn.Keychain,
	desc DependencyDescriptor,
	kpConfig config.KpConfig,
	relocatedImageProvider RelocatedImageProvider,
	differ Differ, cs buildk8s.ClientSet) (hasChanges bool, changes string, err error) {

	var summarizer changeSummarizer
	iDiffer := &ImportDiffer{
		Differ:         differ,
		RelocatedImageProvider: relocatedImageProvider,
	}

	err = writeLifecycleChange(ctx, keychain, kpConfig, desc.Lifecycle, iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	err = writeClusterStoresChange(ctx, keychain, kpConfig, desc.ClusterStores, iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	err = writeClusterStacksChange(ctx, keychain, kpConfig, desc.GetClusterStacks(), iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	err = writeClusterBuildersChange(ctx, desc.GetClusterBuilders(), iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	return summarizer.hasChanges, summarizer.changes.String(), nil
}

type changeSummarizer struct {
	hasChanges bool
	changes    strings.Builder
	diffs      strings.Builder
}

func (cs *changeSummarizer) writeChange(header string) {
	if cs.changes.Len() == 0 {
		cs.changes.WriteString("Changes\n\n")
	}

	cs.changes.WriteString(fmt.Sprintf("%s\n\n", header))

	change := cs.diffs.String()
	if change == "" {
		cs.changes.WriteString("No Changes\n\n")
	} else {
		cs.changes.WriteString(change)
		cs.hasChanges = true
	}

	cs.diffs.Reset()
}

func (cs *changeSummarizer) writeDiff(diff string) error {
	var err error
	if diff != "" {
		_, err = cs.diffs.WriteString(fmt.Sprintf("%s\n\n", diff))
	}
	return err
}
