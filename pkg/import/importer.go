package _import

import (
	"context"
	"path"

	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware-tanzu/kpack-cli/pkg/clusterstack"
	"github.com/vmware-tanzu/kpack-cli/pkg/clusterstore"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

type ImageRelocator interface {
	Relocate(keychain authn.Keychain, src v1.Image, destination string) (string, error)
}

type ImageFetcher interface {
	Fetch(keychain authn.Keychain, image string) (v1.Image, error)
}

type TimestampProvider interface {
	GetTimestamp() string
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	PrintStatus(format string, args ...interface{}) error
}

type Importer struct {
	client              versioned.Interface
	k8sClient           kubernetes.Interface
	printer             Printer
	imageRelocator      ImageRelocator
	imageFetcher        ImageFetcher
	waiter              commands.ResourceWaiter
	clusterStoreFactory *clusterstore.Factory
	clusterStackFactory *clusterstack.Factory
	timestampProvider   TimestampProvider
}

type relocatedDescriptor struct {
	lifecycle       *corev1.ConfigMap
	clusterStores   []*v1alpha1.ClusterStore
	clusterStacks   []*v1alpha1.ClusterStack
	clusterBuilders []*v1alpha1.ClusterBuilder
}

func NewImporter(printer Printer, k8sClient kubernetes.Interface, client versioned.Interface, fetcher ImageFetcher, relocator ImageRelocator, waiter commands.ResourceWaiter, timestampProvider TimestampProvider) *Importer {
	return &Importer{
		imageRelocator:      relocator,
		client:              client,
		k8sClient:           k8sClient,
		printer:             printer,
		waiter:              waiter,
		imageFetcher:        fetcher,
		timestampProvider:   timestampProvider,
		clusterStackFactory: clusterstack.NewFactory(printer, relocator, fetcher),
		clusterStoreFactory: clusterstore.NewFactory(printer, relocator, fetcher),
	}
}

func (i *Importer) ReadDescriptor(rawDescriptor string) (DependencyDescriptor, error) {
	var api API
	if err := yaml.Unmarshal([]byte(rawDescriptor), &api); err != nil {
		return DependencyDescriptor{}, err
	}

	var descriptor DependencyDescriptor
	switch api.Version {
	case APIVersionV1:
		var d1 DependencyDescriptorV1
		if err := yaml.Unmarshal([]byte(rawDescriptor), &d1); err != nil {
			return DependencyDescriptor{}, err
		}
		descriptor = d1.ToNextVersion()
	case CurrentAPIVersion:
		if err := yaml.Unmarshal([]byte(rawDescriptor), &descriptor); err != nil {
			return DependencyDescriptor{}, err
		}
	default:
		return DependencyDescriptor{}, errors.Errorf("did not find expected apiVersion, must be one of: %s", []string{APIVersionV1, CurrentAPIVersion})
	}

	if err := descriptor.Validate(); err != nil {
		return DependencyDescriptor{}, err
	}

	return descriptor, nil
}

func (i *Importer) ImportDescriptor(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, rawDescriptor string) ([]runtime.Object, error) {
	descriptor, err := i.ReadDescriptor(rawDescriptor)
	if err != nil {
		return nil, err
	}

	rDescriptor, objects, err := i.relocateDescriptor(ctx, keychain, kpConfig, i.timestampProvider.GetTimestamp(), descriptor)
	if err != nil {
		return nil, err
	}

	if rDescriptor.lifecycle != nil {
		if err := i.updateLifecycleConfigMap(ctx, rDescriptor.lifecycle); err != nil {
			return nil, err
		}
	}

	storeToGeneration := map[string]int64{}
	for _, store := range rDescriptor.clusterStores {
		gen, err := i.saveClusterStore(ctx, store)
		if err != nil {
			return nil, err
		}

		storeToGeneration[store.Name] = gen
	}

	stackToGeneration := map[string]int64{}
	for _, stack := range rDescriptor.clusterStacks {
		gen, err := i.saveClusterStack(ctx, stack)
		if err != nil {
			return nil, err
		}

		stackToGeneration[stack.Name] = gen
	}

	for _, builder := range rDescriptor.clusterBuilders {
		if err := i.saveClusterBuilder(ctx, storeToGeneration, stackToGeneration, builder); err != nil {
			return nil, err
		}
	}

	return objects, nil
}

func (i *Importer) ImportDescriptorDryRun(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, rawDescriptor string) ([]runtime.Object, error) {
	descriptor, err := i.ReadDescriptor(rawDescriptor)
	if err != nil {
		return nil, err
	}

	_, objects, err := i.relocateDescriptor(ctx, keychain, kpConfig, i.timestampProvider.GetTimestamp(), descriptor)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (i *Importer) relocateDescriptor(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, ts string, descriptor DependencyDescriptor) (relocatedDescriptor, []runtime.Object, error) {
	var (
		updatedLifecycle *corev1.ConfigMap
		err              error
		objs             []runtime.Object
	)

	if descriptor.HasLifecycleImage() {
		updatedLifecycle, err = i.relocateLifecycle(ctx, keychain, kpConfig, ts, descriptor.GetLifecycleImage())
		if err != nil {
			return relocatedDescriptor{}, nil, err
		}
		objs = append(objs, updatedLifecycle)
	}

	clusterstores := make([]*v1alpha1.ClusterStore, 0)
	for _, clusterStore := range descriptor.ClusterStores {
		rStore, err := i.constructClusterStore(ctx, keychain, kpConfig, clusterStore)
		if err != nil {
			return relocatedDescriptor{}, nil, err
		}
		rStore.Annotations = k8s.MergeAnnotations(rStore.Annotations, map[string]string{"kpack.io/import-timestamp": ts})

		clusterstores = append(clusterstores, rStore)
		objs = append(objs, rStore)
	}

	clusterstacks := make([]*v1alpha1.ClusterStack, 0)
	for _, clusterStack := range descriptor.GetClusterStacks() {
		rStack, err := i.constructClusterStack(keychain, kpConfig, clusterStack)
		if err != nil {
			return relocatedDescriptor{}, nil, err
		}
		rStack.Annotations = k8s.MergeAnnotations(rStack.Annotations, map[string]string{"kpack.io/import-timestamp": ts})

		clusterstacks = append(clusterstacks, rStack)
		objs = append(objs, rStack)
	}

	clusterBuilders := make([]*v1alpha1.ClusterBuilder, 0)
	for _, clusterBuilder := range descriptor.GetClusterBuilders() {
		rBuilder, err := i.constructClusterBuilder(kpConfig, clusterBuilder)
		if err != nil {
			return relocatedDescriptor{}, nil, err
		}
		rBuilder.Annotations = k8s.MergeAnnotations(rBuilder.Annotations, map[string]string{"kpack.io/import-timestamp": ts})

		clusterBuilders = append(clusterBuilders, rBuilder)
		objs = append(objs, rBuilder)
	}

	return relocatedDescriptor{
		lifecycle:       updatedLifecycle,
		clusterStores:   clusterstores,
		clusterStacks:   clusterstacks,
		clusterBuilders: clusterBuilders,
	}, objs, nil
}

func (i *Importer) relocateLifecycle(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, ts, lifecyle string) (*corev1.ConfigMap, error) {
	if err := i.printer.PrintStatus("Importing Lifecycle..."); err != nil {
		return nil, err
	}

	lifecycleImage, err := i.imageFetcher.Fetch(keychain, lifecyle)
	if err != nil {
		return nil, err
	}

	relocatedLifecycle, err := i.imageRelocator.Relocate(keychain, lifecycleImage, path.Join(kpConfig.CanonicalRepository, "lifecycle"))
	if err != nil {
		return nil, err
	}

	existingLifecycleConfig, err := i.k8sClient.CoreV1().ConfigMaps("kpack").Get(ctx, "lifecycle-image", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	newConfigMap := existingLifecycleConfig.DeepCopy()

	newConfigMap.SetAnnotations(map[string]string{"kpack.io/import-timestamp": ts})
	newConfigMap.Data["image"] = relocatedLifecycle
	return newConfigMap, nil
}

func (i *Importer) constructClusterStore(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, store ClusterStore) (*v1alpha1.ClusterStore, error) {
	if err := i.printer.PrintStatus("Importing ClusterStore '%s'...", store.Name); err != nil {
		return nil, err
	}

	existingStore, err := i.client.KpackV1alpha1().ClusterStores().Get(ctx, store.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}
	if k8serrors.IsNotFound(err) {
		existingStore = nil
	}

	if existingStore != nil {
		updatedStore := existingStore.DeepCopy()
		updatedStore, _, err = i.clusterStoreFactory.AddToStore(keychain, updatedStore, kpConfig, buildpackagesForSource(store.Sources)...)
		if err != nil {
			return nil, err
		}

		return updatedStore, nil
	}

	newStore, err := i.clusterStoreFactory.MakeStore(keychain, store.Name, kpConfig, buildpackagesForSource(store.Sources)...)
	if err != nil {
		return nil, err
	}
	return newStore, nil
}

func (i *Importer) constructClusterStack(keychain authn.Keychain, kpConfig config.KpConfig, stack ClusterStack) (*v1alpha1.ClusterStack, error) {
	if err := i.printer.PrintStatus("Importing ClusterStack '%s'...", stack.Name); err != nil {
		return nil, err
	}

	newStack, err := i.clusterStackFactory.MakeStack(keychain, stack.Name, stack.BuildImage.Image, stack.RunImage.Image, kpConfig)
	if err != nil {
		return nil, err
	}

	return newStack, nil
}

func (i *Importer) constructClusterBuilder(kpConfig config.KpConfig, builder ClusterBuilder) (*v1alpha1.ClusterBuilder, error) {
	if err := i.printer.PrintStatus("Importing ClusterBuilder '%s'...", builder.Name); err != nil {
		return nil, err
	}

	newCB := &v1alpha1.ClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterBuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        builder.Name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.ClusterBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: path.Join(kpConfig.CanonicalRepository, builder.Name),
				Stack: corev1.ObjectReference{
					Name: builder.ClusterStack,
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: builder.ClusterStore,
					Kind: v1alpha1.ClusterStoreKind,
				},
				Order: builder.Order,
			},
			ServiceAccountRef: kpConfig.ServiceAccount,
		},
	}

	err := k8s.SetLastAppliedCfg(newCB)
	if err != nil {
		return nil, err
	}

	return newCB, nil
}

func (i *Importer) updateLifecycleConfigMap(ctx context.Context, updatedLifecycle *corev1.ConfigMap) error {
	_, err := i.k8sClient.CoreV1().ConfigMaps("kpack").Update(ctx, updatedLifecycle, metav1.UpdateOptions{})

	return err
}

func (i *Importer) saveClusterStore(ctx context.Context, relocatedStore *v1alpha1.ClusterStore) (int64, error) {
	existingStore, err := i.client.KpackV1alpha1().ClusterStores().Get(ctx, relocatedStore.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return 0, err
	}

	var store *v1alpha1.ClusterStore
	if k8serrors.IsNotFound(err) {
		store, err = i.client.KpackV1alpha1().ClusterStores().Create(ctx, relocatedStore, metav1.CreateOptions{})
		if err != nil {
			return 0, err
		}
	} else {
		updateStore := existingStore.DeepCopy()
		updateStore.Spec.Sources = createBuildpackageSuperset(updateStore, relocatedStore)
		updateStore.Annotations = k8s.MergeAnnotations(updateStore.Annotations, relocatedStore.Annotations)
		store, err = i.client.KpackV1alpha1().ClusterStores().Update(ctx, updateStore, metav1.UpdateOptions{})
		if err != nil {
			return 0, err
		}
	}

	if err := i.waiter.Wait(ctx, store); err != nil {
		return 0, err
	}

	return store.Generation, nil
}

func (i *Importer) saveClusterStack(ctx context.Context, relocatedStack *v1alpha1.ClusterStack) (int64, error) {
	exstingStack, err := i.client.KpackV1alpha1().ClusterStacks().Get(ctx, relocatedStack.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return 0, err
	}

	var stack *v1alpha1.ClusterStack
	if k8serrors.IsNotFound(err) {
		stack, err = i.client.KpackV1alpha1().ClusterStacks().Create(ctx, relocatedStack, metav1.CreateOptions{})
		if err != nil {
			return 0, err
		}
	} else {
		updateStack := exstingStack.DeepCopy()
		updateStack.Spec = relocatedStack.Spec
		updateStack.Annotations = k8s.MergeAnnotations(updateStack.Annotations, relocatedStack.Annotations)
		stack, err = i.client.KpackV1alpha1().ClusterStacks().Update(ctx, updateStack, metav1.UpdateOptions{})
		if err != nil {
			return 0, err
		}
	}
	if err := i.waiter.Wait(ctx, stack); err != nil {
		return 0, err
	}

	return stack.Generation, nil
}

func (i *Importer) saveClusterBuilder(ctx context.Context, storeToGeneration, stackToGeneration map[string]int64, relocatedBuilder *v1alpha1.ClusterBuilder) error {
	existingBuilder, err := i.client.KpackV1alpha1().ClusterBuilders().Get(ctx, relocatedBuilder.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	var builder *v1alpha1.ClusterBuilder
	if k8serrors.IsNotFound(err) {
		builder, err = i.client.KpackV1alpha1().ClusterBuilders().Create(ctx, relocatedBuilder, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		updateBuilder := existingBuilder.DeepCopy()
		updateBuilder.Spec = relocatedBuilder.Spec
		updateBuilder.Annotations = k8s.MergeAnnotations(updateBuilder.Annotations, relocatedBuilder.Annotations)
		builder, err = i.client.KpackV1alpha1().ClusterBuilders().Update(ctx, updateBuilder, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return i.waiter.Wait(ctx, builder, builderHasResolved(storeToGeneration[relocatedBuilder.Spec.Store.Name], stackToGeneration[relocatedBuilder.Spec.Stack.Name]))
}

func buildpackagesForSource(sources []Source) []string {
	var buildpackages []string
	for _, s := range sources {
		buildpackages = append(buildpackages, s.Image)
	}
	return buildpackages
}

func createBuildpackageSuperset(firstStore, secondStore *v1alpha1.ClusterStore) []v1alpha1.StoreImage {
	result := firstStore.Spec.Sources

	for _, source := range secondStore.Spec.Sources {
		if !sourcesContainsSourceImage(firstStore.Spec.Sources, source) {
			result = append(result, source)
		}
	}

	return result
}

func sourcesContainsSourceImage(sources []v1alpha1.StoreImage, sourceImage v1alpha1.StoreImage) bool {
	for _, source := range sources {
		if source == sourceImage {
			return true
		}
	}

	return false
}
