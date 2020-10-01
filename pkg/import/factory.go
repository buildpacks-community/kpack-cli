package _import

import (
	"encoding/json"
	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	k8s "github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"path"
)

const (
	importNamespace          = "kpack"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
	importTimestampKey       = "kpack.io/import-timestamp"
)

type TimestampProvider interface {
	GetTimestamp() string
}

type Factory struct {
	Client            kpack.Interface
	TimestampProvider TimestampProvider
	objects           []runtime.Object
	CommandHelper     *commands.CommandHelper
}

func (f *Factory) Objects() []runtime.Object {
	return f.objects
}

func (f *Factory) ImportClusterStores(clusterStores []ClusterStore, factory *clusterstore.Factory, repository string) error {
	for _, store := range clusterStores {
		f.CommandHelper.Printlnf("Importing Cluster Store '%s'...", store.Name)

		var buildpackages []string
		for _, s := range store.Sources {
			buildpackages = append(buildpackages, s.Image)
		}

		curStore, err := f.Client.KpackV1alpha1().ClusterStores().Get(store.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			newStore, err := factory.MakeStore(store.Name, buildpackages...)
			if err != nil {
				return err
			}

			newStore.Annotations[importTimestampKey] = f.TimestampProvider.GetTimestamp()

			if !f.CommandHelper.IsDryRun() {
				if newStore, err = f.Client.KpackV1alpha1().ClusterStores().Create(newStore); err != nil {
					return err
				}
			}
			f.trackObj(newStore)
		} else {
			updatedStore, _, err := factory.AddToStore(curStore, repository, buildpackages...)
			if err != nil {
				return err
			}

			curStore.Annotations = k8s.MergeAnnotations(curStore.Annotations, map[string]string{importTimestampKey: f.TimestampProvider.GetTimestamp()})

			if !f.CommandHelper.IsDryRun() {
				if updatedStore, err = f.Client.KpackV1alpha1().ClusterStores().Update(updatedStore); err != nil {
					return err
				}
			}
			f.trackObj(updatedStore)
		}
	}
	return nil
}

func (f *Factory) ImportClusterStacks(clusterStacks []ClusterStack, factory *clusterstack.Factory) error {
	for _, stack := range clusterStacks {
		f.CommandHelper.Printlnf("Importing Cluster Stack '%s'...", stack.Name)

		factory.Printer = f.CommandHelper
		factory.BuildImageRef = stack.BuildImage.Image // FIXME
		factory.RunImageRef = stack.RunImage.Image     // FIXME

		newStack, err := factory.MakeStack(stack.Name)
		if err != nil {
			return err
		}

		newStack.Annotations[importTimestampKey] = f.TimestampProvider.GetTimestamp()

		curStack, err := f.Client.KpackV1alpha1().ClusterStacks().Get(stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			if !f.CommandHelper.IsDryRun() {
				if newStack, err = f.Client.KpackV1alpha1().ClusterStacks().Create(newStack); err != nil {
					return err
				}
			}
			f.trackObj(newStack)
		} else {
			updateStack := curStack.DeepCopy()
			updateStack.Spec = newStack.Spec
			updateStack.Annotations = k8s.MergeAnnotations(updateStack.Annotations, newStack.Annotations)

			if !f.CommandHelper.IsDryRun() {
				if updateStack, err = f.Client.KpackV1alpha1().ClusterStacks().Update(updateStack); err != nil {
					return err
				}
			}
			f.trackObj(updateStack)
		}
	}
	return nil
}

func (f *Factory) ImportClusterBuilders(clusterBuilders []ClusterBuilder, repository string, sa string) error {
	for _, ccb := range clusterBuilders {
		if err := f.CommandHelper.Printlnf("Importing Cluster Builder '%s'...", ccb.Name); err != nil {
			return err
		}

		newCB, err := f.makeClusterBuilder(ccb, repository, sa)
		if err != nil {
			return err
		}

		newCB.Annotations[importTimestampKey] = f.TimestampProvider.GetTimestamp()

		curCCB, err := f.Client.KpackV1alpha1().ClusterBuilders().Get(ccb.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			if !f.CommandHelper.IsDryRun() {
				if newCB, err = f.Client.KpackV1alpha1().ClusterBuilders().Create(newCB); err != nil {
					return err
				}
			}
			f.trackObj(newCB)
		} else {
			updateCB := curCCB.DeepCopy()
			updateCB.Spec = newCB.Spec
			updateCB.Annotations = k8s.MergeAnnotations(updateCB.Annotations, newCB.Annotations)

			if !f.CommandHelper.IsDryRun() {
				if updateCB, err = f.Client.KpackV1alpha1().ClusterBuilders().Update(updateCB); err != nil {
					return err
				}
			}
			f.trackObj(updateCB)
		}
	}
	return nil
}

func (f *Factory) trackObj(obj runtime.Object) {
	f.objects = append(f.objects, obj)
}

func (f Factory) makeClusterBuilder(ccb ClusterBuilder, repository string, sa string) (*v1alpha1.ClusterBuilder, error) {
	newCCB := &v1alpha1.ClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterBuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ccb.Name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.ClusterBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: path.Join(repository, ccb.Name),
				Stack: corev1.ObjectReference{
					Name: ccb.ClusterStack,
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: ccb.ClusterStore,
					Kind: v1alpha1.ClusterStoreKind,
				},
				Order: ccb.Order,
			},
		},
	}

	if sa != "" {
		newCCB.Spec.ServiceAccountRef = corev1.ObjectReference{
			Namespace: importNamespace,
			Name:      sa,
		}
	}

	marshal, err := json.Marshal(newCCB)
	if err != nil {
		return nil, err
	}
	newCCB.Annotations[kubectlLastAppliedConfig] = string(marshal)

	return newCCB, nil
}
