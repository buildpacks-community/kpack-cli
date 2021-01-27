package _import

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	buildk8s "github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/lifecycle"
)

type changeWriter interface {
	writeDiff(diffs string) error
	writeChange(header string)
}

func writeLifecycleChange(newLifecycle Lifecycle, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	if newLifecycle.Image != "" {
		oldImg, err := lifecycle.GetImage(cs.K8sClient)
		if err != nil {
			return err
		}

		diff, err := differ.DiffLifecycle(oldImg, newLifecycle.Image)
		if err != nil {
			return err
		}

		if err = cw.writeDiff(diff); err != nil {
			return err
		}
	}

	cw.writeChange("Lifecycle")
	return nil
}

func writeClusterStoresChange(stores []ClusterStore, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, store := range stores {
		oldStore, err := cs.KpackClient.KpackV1alpha1().ClusterStores().Get(store.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldStore = nil
		}

		diff, err := differ.DiffClusterStore(oldStore, store)
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

func writeClusterStacksChange(stacks []ClusterStack, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, stack := range stacks {
		oldStack, err := cs.KpackClient.KpackV1alpha1().ClusterStacks().Get(stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if k8serrors.IsNotFound(err) {
			oldStack = nil
		}

		diff, err := differ.DiffClusterStack(oldStack, stack)
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

func writeClusterBuildersChange(builders []ClusterBuilder, differ *ImportDiffer, cs buildk8s.ClientSet, cw changeWriter) error {
	for _, builder := range builders {
		oldBuilder, err := cs.KpackClient.KpackV1alpha1().ClusterBuilders().Get(builder.Name, metav1.GetOptions{})
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
