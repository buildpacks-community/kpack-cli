package _import

import (
	"fmt"
	"strings"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	buildk8s "github.com/pivotal/build-service-cli/pkg/k8s"
)

func SummarizeChange(
	desc DependencyDescriptor,
	storeFactory *clusterstore.Factory, stackFactory *clusterstack.Factory,
	differ Differ, cs buildk8s.ClientSet) (hasChanges bool, changes string, err error) {

	var summarizer changeSummarizer
	iDiffer := &ImportDiffer{
		Differ:         differ,
		StoreRefGetter: storeFactory,
		StackRefGetter: stackFactory,
	}

	err = writeLifecycleChange(desc.Lifecycle, iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	err = writeClusterStoresChange(desc.ClusterStores, iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	err = writeClusterStacksChange(desc.GetClusterStacks(), iDiffer, cs, &summarizer)
	if err != nil {
		return
	}

	err = writeClusterBuildersChange(desc.GetClusterBuilders(), iDiffer, cs, &summarizer)
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
