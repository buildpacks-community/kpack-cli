package _import

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	kpacktesthelpers "github.com/pivotal/kpack/pkg/reconciler/testhelpers"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
	watchTools "k8s.io/client-go/tools/watch"

	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestImporter(t *testing.T) {
	spec.Run(t, "TestImporter", testImporter)
}

func testImporter(t *testing.T, when spec.G, it spec.S) {
	var (
		lifecycleDigest  = "lifecycledigest"
		dotnetCoreDigest = "dotnetcoredigest"
		dotnetCoreId     = "dotnet/core"
		stackId          = "io.stacks.mycoolstack"
		buildImageDigest = "buildimagedigest"
		runImageDigest   = "runimagedigest"

		kpConfig = config.NewKpConfig(
			"gcr.io/my-cool-repo",
			corev1.ObjectReference{
				Namespace: "some-namespace",
				Name:      "some-serviceaccount",
			},
		)

		existingLifecycle = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lifecycle-image",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"image": "old/image",
			},
		}
	)
	when("importing dependencies", func() {
		var (
			expectedDefaultClusterStore   runtime.Object
			expectedClusterStack          runtime.Object
			expectedDefaultClusterStack   runtime.Object
			expectedClusterBuilder        runtime.Object
			expectedDefaultClusterBuilder runtime.Object
		)

		it.Before(func() {
			expectedDefaultClusterStore = annotate(t, &v1alpha2.ClusterStore{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterStore",
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
				Spec: v1alpha2.ClusterStoreSpec{
					Sources: []corev1alpha1.StoreImage{
						{Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", dotnetCoreDigest)},
					},
					ServiceAccountRef: &corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}, kubectlAnnotation, timestampAnnotation)
			expectedClusterStack = annotate(t, &v1alpha2.ClusterStack{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterStack",
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "base",
				},
				Spec: v1alpha2.ClusterStackSpec{
					Id: stackId,
					BuildImage: v1alpha2.ClusterStackSpecImage{
						Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", buildImageDigest),
					},
					RunImage: v1alpha2.ClusterStackSpecImage{
						Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", runImageDigest),
					},
					ServiceAccountRef: &corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}, timestampAnnotation)
			expectedDefaultClusterStack = annotate(t, &v1alpha2.ClusterStack{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterStack",
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
				Spec: v1alpha2.ClusterStackSpec{
					Id: stackId,
					BuildImage: v1alpha2.ClusterStackSpecImage{
						Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", buildImageDigest),
					},
					RunImage: v1alpha2.ClusterStackSpecImage{
						Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", runImageDigest),
					},
					ServiceAccountRef: &corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}, timestampAnnotation)
			expectedClusterBuilder = annotate(t, &v1alpha2.ClusterBuilder{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterBuilder",
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "base",
				},
				Spec: v1alpha2.ClusterBuilderSpec{
					BuilderSpec: v1alpha2.BuilderSpec{
						Tag: "gcr.io/my-cool-repo:clusterbuilder-base",
						Stack: corev1.ObjectReference{
							Kind: "ClusterStack",
							Name: "base",
						},
						Store: corev1.ObjectReference{
							Kind: "ClusterStore",
							Name: "default",
						},
						Order: []corev1alpha1.OrderEntry{
							{
								[]corev1alpha1.BuildpackRef{
									{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "tanzu-buildpacks/dotnet-core",
										},
										Optional: false,
									},
								},
							},
						},
					},
					ServiceAccountRef: corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}, kubectlAnnotation, timestampAnnotation)
			expectedDefaultClusterBuilder = annotate(t, &v1alpha2.ClusterBuilder{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterBuilder",
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
				Spec: v1alpha2.ClusterBuilderSpec{
					BuilderSpec: v1alpha2.BuilderSpec{
						Tag: "gcr.io/my-cool-repo:clusterbuilder-default",
						Stack: corev1.ObjectReference{
							Kind: "ClusterStack",
							Name: "base",
						},
						Store: corev1.ObjectReference{
							Kind: "ClusterStore",
							Name: "default",
						},
						Order: []corev1alpha1.OrderEntry{
							{
								[]corev1alpha1.BuildpackRef{
									{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "tanzu-buildpacks/dotnet-core",
										},
										Optional: false,
									},
								},
							},
						},
					},
					ServiceAccountRef: corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
					},
				},
			}, kubectlAnnotation, timestampAnnotation)
		})

		it("can import on a new cluster", func() {
			TestImport{
				Images: map[string]v1.Image{
					"new-image.com/lifecycle":              fakes.NewFakeImage(lifecycleDigest),
					"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), dotnetCoreDigest),
					"new-image.com/stacks/base/run":        fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, runImageDigest),
					"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, buildImageDigest),
				},
				Objects: []runtime.Object{
					existingLifecycle,
				},
				KpConfig: kpConfig,
				DependencyDescriptor: `
apiVersion: kp.kpack.io/v1alpha3
kind: DependencyDescriptor
defaultClusterBuilder: base
defaultClusterStack: base
lifecycle:
  image: new-image.com/lifecycle
clusterStores:
- name: default
  sources:
  - image: new-image.com/buildpacks/dotnet-core
clusterStacks:
- name: base
  buildImage:
    image: new-image.com/stacks/base/build
  runImage:
    image: new-image.com/stacks/base/run
clusterBuilders:
- name: base
  clusterStack: base
  clusterStore: default
  order:
  - group:
    - id: tanzu-buildpacks/dotnet-core
`,
				ExpectCreates: []runtime.Object{
					expectedDefaultClusterStore,
					expectedClusterStack,
					expectedDefaultClusterStack,
					expectedClusterBuilder,
					expectedDefaultClusterBuilder,
				},
				ExpectPatches: []string{
					`{"data":{"image":"gcr.io/my-cool-repo@sha256:lifecycledigest"},"metadata":{"annotations":{"kpack.io/import-timestamp":"0001-01-01 00:00:00 +0000 UTC"}}}`,
				},
			}.TestImporter(t)
		})

		it("can import v1alpha1 descriptor on new cluster", func() {
			dotnetCoreDigest := "dotnetcoredigest"
			dotnetCoreId := "dotnet/core"
			stackId := "io.stacks.mycoolstack"
			buildImageDigest := "buildimagedigest"
			runImageDigest := "runimagedigest"

			TestImport{
				Images: map[string]v1.Image{
					"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), dotnetCoreDigest),
					"new-image.com/stacks/base/run":        fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, runImageDigest),
					"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, buildImageDigest),
				},
				Objects: []runtime.Object{
					existingLifecycle,
				},
				KpConfig: kpConfig,
				DependencyDescriptor: `
apiVersion: kp.kpack.io/v1alpha1
kind: DependencyDescriptor
defaultClusterBuilder: base
defaultStack: base
stores:
- name: default
  sources:
  - image: new-image.com/buildpacks/dotnet-core
stacks:
- name: base
  buildImage:
   image: new-image.com/stacks/base/build
  runImage:
   image: new-image.com/stacks/base/run
clusterBuilders:
- name: base
  stack: base
  store: default
  order:
  - group:
    - id: tanzu-buildpacks/dotnet-core
`,
				ExpectCreates: []runtime.Object{
					expectedDefaultClusterStore,
					expectedClusterStack,
					expectedDefaultClusterStack,
					expectedClusterBuilder,
					expectedDefaultClusterBuilder,
				},
			}.TestImporter(t)
		})

		when("importing to an existing cluster", func() {
			var (
				existingClusterStore          runtime.Object
				existingClusterStack          runtime.Object
				existingDefaultClusterStack   runtime.Object
				existingClusterBuilder        runtime.Object
				existingDefaultClusterBuilder runtime.Object
			)

			it.Before(func() {
				existingClusterStore = annotate(t, &v1alpha2.ClusterStore{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterStore",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
					Spec: v1alpha2.ClusterStoreSpec{
						Sources: []corev1alpha1.StoreImage{
							{Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", dotnetCoreDigest)},
						},
						ServiceAccountRef: &corev1.ObjectReference{
							Namespace: "some-namespace",
							Name:      "some-serviceaccount",
						},
					},
				}, timestampAnnotation)
				existingClusterStack = annotate(t, &v1alpha2.ClusterStack{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterStack",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "base",
					},
					Spec: v1alpha2.ClusterStackSpec{
						Id: stackId,
						BuildImage: v1alpha2.ClusterStackSpecImage{
							Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", buildImageDigest),
						},
						RunImage: v1alpha2.ClusterStackSpecImage{
							Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", runImageDigest),
						},
						ServiceAccountRef: &corev1.ObjectReference{
							Namespace: "some-namespace",
							Name:      "some-serviceaccount",
						},
					},
					Status: v1alpha2.ClusterStackStatus{
						Status: corev1alpha1.Status{},
						ResolvedClusterStack: v1alpha2.ResolvedClusterStack{
							BuildImage: v1alpha2.ClusterStackStatusImage{
								LatestImage: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", buildImageDigest),
								Image:       "",
							},
							RunImage: v1alpha2.ClusterStackStatusImage{
								LatestImage: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", runImageDigest),
								Image:       "",
							},
						},
					},
				}, timestampAnnotation)
				existingDefaultClusterStack = annotate(t, &v1alpha2.ClusterStack{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterStack",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
					Spec: v1alpha2.ClusterStackSpec{
						Id: stackId,
						BuildImage: v1alpha2.ClusterStackSpecImage{
							Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", buildImageDigest),
						},
						RunImage: v1alpha2.ClusterStackSpecImage{
							Image: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", runImageDigest),
						},
						ServiceAccountRef: &corev1.ObjectReference{
							Namespace: "some-namespace",
							Name:      "some-serviceaccount",
						},
					},
					Status: v1alpha2.ClusterStackStatus{
						Status: corev1alpha1.Status{},
						ResolvedClusterStack: v1alpha2.ResolvedClusterStack{
							BuildImage: v1alpha2.ClusterStackStatusImage{
								LatestImage: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", buildImageDigest),
								Image:       "",
							},
							RunImage: v1alpha2.ClusterStackStatusImage{
								LatestImage: fmt.Sprintf("gcr.io/my-cool-repo@sha256:%s", runImageDigest),
								Image:       "",
							},
						},
					},
				}, timestampAnnotation)
				existingClusterBuilder = annotate(t, &v1alpha2.ClusterBuilder{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterBuilder",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "base",
					},
					Spec: v1alpha2.ClusterBuilderSpec{
						BuilderSpec: v1alpha2.BuilderSpec{
							Tag: "gcr.io/my-cool-repo",
							Stack: corev1.ObjectReference{
								Kind: "ClusterStack",
								Name: "base",
							},
							Store: corev1.ObjectReference{
								Kind: "ClusterStore",
								Name: "default",
							},
							Order: []corev1alpha1.OrderEntry{
								{
									[]corev1alpha1.BuildpackRef{
										{
											BuildpackInfo: corev1alpha1.BuildpackInfo{
												Id: "tanzu-buildpacks/dotnet-core",
											},
											Optional: false,
										},
									},
								},
							},
						},
						ServiceAccountRef: corev1.ObjectReference{
							Namespace: "some-namespace",
							Name:      "some-serviceaccount",
						},
					},
				}, kubectlAnnotation, timestampAnnotation)
				existingDefaultClusterBuilder = annotate(t, &v1alpha2.ClusterBuilder{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterBuilder",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
					Spec: v1alpha2.ClusterBuilderSpec{
						BuilderSpec: v1alpha2.BuilderSpec{
							Tag: "gcr.io/my-cool-repo",
							Stack: corev1.ObjectReference{
								Kind: "ClusterStack",
								Name: "base",
							},
							Store: corev1.ObjectReference{
								Kind: "ClusterStore",
								Name: "default",
							},
							Order: []corev1alpha1.OrderEntry{
								{
									[]corev1alpha1.BuildpackRef{
										{
											BuildpackInfo: corev1alpha1.BuildpackInfo{
												Id: "tanzu-buildpacks/dotnet-core",
											},
											Optional: false,
										},
									},
								},
							},
						},
						ServiceAccountRef: corev1.ObjectReference{
							Namespace: "some-namespace",
							Name:      "some-serviceaccount",
						},
					},
				}, kubectlAnnotation, timestampAnnotation)
			})

			it("can import on an existing cluster", func() {
				newLifecycleDigest := "newlifecycledigest"
				newDotnetCoreDigest := "newdotnetcoredigest"
				newBuildImageDigest := "newbuildimagedigest"
				newRunImageDigest := "newrunimagedigest"
				nodejsDigest := "nodejsdigest"
				nodejsId := "node/js"

				TestImport{
					Images: map[string]v1.Image{
						"new-image.com/lifecycle":              fakes.NewFakeImage(newLifecycleDigest),
						"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), newDotnetCoreDigest),
						"new-image.com/buildpacks/nodejs":      fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", nodejsId), nodejsDigest),
						"new-image.com/stacks/base/run":        fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, newRunImageDigest),
						"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, newBuildImageDigest),
					},
					Objects: []runtime.Object{
						existingLifecycle,
						existingClusterStore,
						existingClusterStack,
						existingDefaultClusterStack,
						existingClusterBuilder,
						existingDefaultClusterBuilder,
					},
					KpConfig: kpConfig,
					DependencyDescriptor: `
apiVersion: kp.kpack.io/v1alpha3
kind: DependencyDescriptor
defaultClusterBuilder: base
defaultClusterStack: base
lifecycle:
  image: new-image.com/lifecycle
clusterStores:
- name: default
  sources:
  - image: new-image.com/buildpacks/dotnet-core
  - image: new-image.com/buildpacks/nodejs
clusterStacks:
- name: base
  buildImage:
    image: new-image.com/stacks/base/build
  runImage:
    image: new-image.com/stacks/base/run
clusterBuilders:
- name: base
  clusterStack: base
  clusterStore: default
  order:
  - group:
    - id: tanzu-buildpacks/dotnet-core
  - group:
    - id: tanzu-buildpacks/nodejs
`,
					ExpectPatches: []string{
						`{"spec":{"sources":[{"image":"gcr.io/my-cool-repo@sha256:dotnetcoredigest"},{"image":"gcr.io/my-cool-repo@sha256:newdotnetcoredigest"},{"image":"gcr.io/my-cool-repo@sha256:nodejsdigest"}]}}`,
						`{"spec":{"buildImage":{"image":"gcr.io/my-cool-repo@sha256:newbuildimagedigest"},"runImage":{"image":"gcr.io/my-cool-repo@sha256:newrunimagedigest"}}}`,
						`{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"base\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"gcr.io/my-cool-repo:clusterbuilder-base\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"base\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"default\"},\"order\":[{\"group\":[{\"id\":\"tanzu-buildpacks/dotnet-core\"}]},{\"group\":[{\"id\":\"tanzu-buildpacks/nodejs\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"}},"spec":{"order":[{"group":[{"id":"tanzu-buildpacks/dotnet-core"}]},{"group":[{"id":"tanzu-buildpacks/nodejs"}]}],"tag":"gcr.io/my-cool-repo:clusterbuilder-base"}}`,
						`{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"gcr.io/my-cool-repo:clusterbuilder-default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"base\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"default\"},\"order\":[{\"group\":[{\"id\":\"tanzu-buildpacks/dotnet-core\"}]},{\"group\":[{\"id\":\"tanzu-buildpacks/nodejs\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"}},"spec":{"order":[{"group":[{"id":"tanzu-buildpacks/dotnet-core"}]},{"group":[{"id":"tanzu-buildpacks/nodejs"}]}],"tag":"gcr.io/my-cool-repo:clusterbuilder-default"}}`,
						`{"data":{"image":"gcr.io/my-cool-repo@sha256:newlifecycledigest"},"metadata":{"annotations":{"kpack.io/import-timestamp":"0001-01-01 00:00:00 +0000 UTC"}}}`,
					},
				}.TestImporter(t)
			})

			it("can import v1alpha1 descriptor on an existing cluster", func() {
				newLifecycleDigest := "newlifecycledigest"
				newDotnetCoreDigest := "newdotnetcoredigest"
				newBuildImageDigest := "newbuildimagedigest"
				newRunImageDigest := "newrunimagedigest"
				nodejsDigest := "nodejsdigest"
				nodejsId := "node/js"

				TestImport{
					Images: map[string]v1.Image{
						"new-image.com/lifecycle":              fakes.NewFakeImage(newLifecycleDigest),
						"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), newDotnetCoreDigest),
						"new-image.com/buildpacks/nodejs":      fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", nodejsId), nodejsDigest),
						"new-image.com/stacks/base/run":        fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, newRunImageDigest),
						"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, newBuildImageDigest),
					},
					Objects: []runtime.Object{
						existingLifecycle,
						existingClusterStore,
						existingClusterStack,
						existingDefaultClusterStack,
						existingClusterBuilder,
						existingDefaultClusterBuilder,
					},
					KpConfig: kpConfig,
					DependencyDescriptor: `
apiVersion: kp.kpack.io/v1alpha1
kind: DependencyDescriptor
defaultClusterBuilder: base
defaultStack: base
stores:
- name: default
  sources:
  - image: new-image.com/buildpacks/dotnet-core
  - image: new-image.com/buildpacks/nodejs
stacks:
- name: base
  buildImage:
   image: new-image.com/stacks/base/build
  runImage:
   image: new-image.com/stacks/base/run
clusterBuilders:
- name: base
  stack: base
  store: default
  order:
  - group:
    - id: tanzu-buildpacks/dotnet-core
  - group:
    - id: tanzu-buildpacks/nodejs
`,
					ExpectPatches: []string{
						`{"spec":{"buildImage":{"image":"gcr.io/my-cool-repo@sha256:newbuildimagedigest"},"runImage":{"image":"gcr.io/my-cool-repo@sha256:newrunimagedigest"}}}`,
						`{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"base\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"gcr.io/my-cool-repo:clusterbuilder-base\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"base\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"default\"},\"order\":[{\"group\":[{\"id\":\"tanzu-buildpacks/dotnet-core\"}]},{\"group\":[{\"id\":\"tanzu-buildpacks/nodejs\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"}},"spec":{"order":[{"group":[{"id":"tanzu-buildpacks/dotnet-core"}]},{"group":[{"id":"tanzu-buildpacks/nodejs"}]}],"tag":"gcr.io/my-cool-repo:clusterbuilder-base"}}`,
						`{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"kind\":\"ClusterBuilder\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"default\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"gcr.io/my-cool-repo:clusterbuilder-default\",\"stack\":{\"kind\":\"ClusterStack\",\"name\":\"base\"},\"store\":{\"kind\":\"ClusterStore\",\"name\":\"default\"},\"order\":[{\"group\":[{\"id\":\"tanzu-buildpacks/dotnet-core\"}]},{\"group\":[{\"id\":\"tanzu-buildpacks/nodejs\"}]}],\"serviceAccountRef\":{\"namespace\":\"some-namespace\",\"name\":\"some-serviceaccount\"}},\"status\":{\"stack\":{}}}"}},"spec":{"order":[{"group":[{"id":"tanzu-buildpacks/dotnet-core"}]},{"group":[{"id":"tanzu-buildpacks/nodejs"}]}],"tag":"gcr.io/my-cool-repo:clusterbuilder-default"}}`,
						`{"spec":{"sources":[{"image":"gcr.io/my-cool-repo@sha256:dotnetcoredigest"},{"image":"gcr.io/my-cool-repo@sha256:newdotnetcoredigest"},{"image":"gcr.io/my-cool-repo@sha256:nodejsdigest"}]}}`,
					},
				}.TestImporter(t)
			})
		})

		it("does not create any resources if any relocation fails", func() {
			TestImport{
				Images: map[string]v1.Image{
					"new-image.com/lifecycle":              fakes.NewFakeImage(lifecycleDigest),
					"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), dotnetCoreDigest),
					"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, buildImageDigest),
				},
				Objects: []runtime.Object{
					existingLifecycle,
				},
				KpConfig: kpConfig,
				DependencyDescriptor: `
apiVersion: kp.kpack.io/v1alpha3
kind: DependencyDescriptor
defaultClusterBuilder: base
defaultClusterStack: base
lifecycle:
  image: new-image.com/lifecycle
clusterStores:
- name: default
  sources:
  - image: new-image.com/buildpacks/dotnet-core
clusterStacks:
- name: base
  buildImage:
    image: new-image.com/stacks/base/build
  runImage:
    image: new-image.com/stacks/base/run
clusterBuilders:
- name: base
  clusterStack: base
  clusterStore: default
  order:
  - group:
    - id: tanzu-buildpacks/dotnet-core
`,

				ExpectErr: errors.New("buddy we don't have your image, check another registry"),
			}.TestImporter(t)
		})
	})

	when("importing with the dry run", func() {
		it("uploads does not create or update any resources", func() {
			TestImport{
				Images: map[string]v1.Image{
					"new-image.com/lifecycle":              fakes.NewFakeImage(lifecycleDigest),
					"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), dotnetCoreDigest),
					"new-image.com/stacks/base/run":        fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, runImageDigest),
					"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, buildImageDigest),
				},
				Objects: []runtime.Object{
					existingLifecycle,
				},
				KpConfig: kpConfig,
				DependencyDescriptor: `
apiVersion: kp.kpack.io/v1alpha3
kind: DependencyDescriptor
defaultClusterBuilder: base
defaultClusterStack: base
lifecycle:
  image: new-image.com/lifecycle
clusterStores:
- name: default
  sources:
  - image: new-image.com/buildpacks/dotnet-core
clusterStacks:
- name: base
  buildImage:
    image: new-image.com/stacks/base/build
  runImage:
    image: new-image.com/stacks/base/run
clusterBuilders:
- name: base
  clusterStack: base
  clusterStore: default
  order:
  - group:
    - id: tanzu-buildpacks/dotnet-core
`,

				DryRun: true,
			}.TestImporter(t)
		})
	})
}

func annotate(t *testing.T, object k8s.Annotatable, f ...func(t *testing.T, object k8s.Annotatable) k8s.Annotatable) runtime.Object {
	for _, function := range f {
		object = function(t, object)
	}
	return object
}

func kubectlAnnotation(t *testing.T, object k8s.Annotatable) k8s.Annotatable {
	err := k8s.SetLastAppliedCfg(object)
	require.NoError(t, err)

	return object
}

func timestampAnnotation(t *testing.T, object k8s.Annotatable) k8s.Annotatable {
	annotations := k8s.MergeAnnotations(object.GetAnnotations(), map[string]string{
		"kpack.io/import-timestamp": time.Time{}.String(),
	})
	object.SetAnnotations(annotations)

	return object
}

type TestImport struct {
	Objects              []runtime.Object
	KpConfig             config.KpConfig
	DependencyDescriptor string
	DryRun               bool
	ExpectUpdates        []clientgotesting.UpdateActionImpl
	ExpectPatches        []string
	Images               map[string]v1.Image
	ExpectCreates        []runtime.Object
	ExpectErr            error
}

func (i TestImport) TestImporter(t *testing.T) {
	t.Helper()
	listers := kpacktesthelpers.NewListers(i.Objects)

	client := kpackfakes.NewSimpleClientset(listers.BuildServiceObjects()...)
	k8sClient := k8sfakes.NewSimpleClientset(listers.GetKubeObjects()...)

	buffer := &bytes.Buffer{}
	var err error
	importer := NewImporter(testLogger{writer: buffer}, k8sClient, client, &fakeFetcher{Images: i.Images}, &fakeRelocator{}, &fakeWaiter{}, &fakeTimestampProvider{ts: time.Time{}.String()})
	if i.DryRun {
		_, err = importer.ImportDescriptorDryRun(context.Background(), authn.NewMultiKeychain(), i.KpConfig, i.DependencyDescriptor)
	} else {
		_, err = importer.ImportDescriptor(context.Background(), authn.NewMultiKeychain(), i.KpConfig, i.DependencyDescriptor)
	}

	if i.ExpectErr != nil {
		assert.EqualError(t, err, i.ExpectErr.Error())
	} else {
		require.NoError(t, err)
	}

	testhelpers.TestK8sAndKpackActions(
		t,
		k8sClient,
		client,
		i.ExpectUpdates,
		i.ExpectCreates,
		nil,
		i.ExpectPatches,
	)
}

type testLogger struct {
	writer io.Writer
}

func (t testLogger) Printlnf(format string, args ...interface{}) error {
	return nil
}

func (t testLogger) PrintStatus(format string, args ...interface{}) error {
	_, err := t.writer.Write([]byte(fmt.Sprintf(format, args...)))
	return err
}

func (t testLogger) Writer() io.Writer {
	return t.writer
}

type fakeFetcher struct {
	Images map[string]v1.Image
}

func (f *fakeFetcher) Fetch(keychain authn.Keychain, image string) (v1.Image, error) {
	img, ok := f.Images[image]
	if !ok {
		return nil, errors.New("buddy we don't have your image, check another registry")
	}

	return img, nil
}

type fakeRelocator struct{}

func (f *fakeRelocator) Relocate(keychain authn.Keychain, src v1.Image, destination string) (string, error) {
	digest, err := src.Digest()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@%s", destination, digest), nil
}

type fakeWaiter struct{}

func (f *fakeWaiter) Wait(ctx context.Context, object runtime.Object, extraChecks ...watchTools.ConditionFunc) error {
	return nil
}

type fakeTimestampProvider struct {
	ts string
}

func (f fakeTimestampProvider) GetTimestamp() string {
	return f.ts
}
