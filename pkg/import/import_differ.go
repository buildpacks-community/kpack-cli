package _import

import (
	"context"
	"sync"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"golang.org/x/sync/errgroup"
)

type StoreRefGetter interface {
	RelocatedBuildpackage(string) (string, error)
}
type StackRefGetter interface {
	RelocatedBuildImage(string) (string, error)
	RelocatedRunImage(string) (string, error)
}

type Differ interface {
	Diff(dOld, dNew interface{}) (string, error)
}

type ImportDiffer struct {
	Differ         Differ
	StoreRefGetter StoreRefGetter
	StackRefGetter StackRefGetter
}

func (id *ImportDiffer) DiffClusterStore(oldCS *v1alpha1.ClusterStore, newCS ClusterStore) (string, error) {
	type void struct{}
	newBPs := map[string]void{}
	mux := &sync.Mutex{}
	errs, _ := errgroup.WithContext(context.Background())

	for _, bp := range newCS.Sources {
		image := bp.Image
		errs.Go(func() error {
			relocatedBP, err := id.StoreRefGetter.RelocatedBuildpackage(image)
			if err != nil {
				return err
			}
			mux.Lock()
			newBPs[relocatedBP] = void{}
			mux.Unlock()
			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return "", err
	}

	oldCSStr := ""
	if oldCS != nil {
		for _, s := range oldCS.Spec.Sources {
			delete(newBPs, s.Image)
		}
		oldCSStr = `Name: default
Sources:`
	}

	if len(newBPs) == 0 {
		return "", nil
	}

	newCS.Sources = []Source{}
	for img := range newBPs {
		newCS.Sources = append(newCS.Sources, Source{Image: img})
	}

	return id.Differ.Diff(oldCSStr, newCS)
}

func (id *ImportDiffer) DiffClusterStack(oldCS *v1alpha1.ClusterStack, newCS ClusterStack) (diff string, err error) {
	newCS.BuildImage.Image, err = id.StackRefGetter.RelocatedBuildImage(newCS.BuildImage.Image)
	if err != nil {
		return "", err
	}
	newCS.RunImage.Image, err = id.StackRefGetter.RelocatedRunImage(newCS.RunImage.Image)
	if err != nil {
		return "", err
	}

	var oldDiffableStack interface{}
	if oldCS != nil {
		oldDiffableStack = ClusterStack{
			Name:       oldCS.Name,
			BuildImage: Source{Image: oldCS.Spec.BuildImage.Image},
			RunImage:   Source{Image: oldCS.Spec.RunImage.Image},
		}
	}

	return id.Differ.Diff(oldDiffableStack, newCS)
}

func (id *ImportDiffer) DiffClusterBuilder(oldCB *v1alpha1.ClusterBuilder, newCB ClusterBuilder) (string, error) {
	var oldDiffableCB interface{}
	if oldCB != nil {
		oldDiffableCB = ClusterBuilder{
			Name:         oldCB.Name,
			ClusterStack: oldCB.Spec.Stack.Name,
			ClusterStore: oldCB.Spec.Store.Name,
			Order:        oldCB.Spec.Order,
		}
	}

	return id.Differ.Diff(oldDiffableCB, newCB)
}
