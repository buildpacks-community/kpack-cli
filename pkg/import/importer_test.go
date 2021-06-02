package _import

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
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
	)
	when("importing dependencies", func() {
		it("can import on a new cluster", func() {
			TestImport{
				Images: map[string]v1.Image{
					"new-image.com/lifecycle":              fakes.NewFakeImage(lifecycleDigest),
					"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), dotnetCoreDigest),
					"new-image.com/stacks/base/run":        fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, runImageDigest),
					"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, buildImageDigest),
				},
				K8sObjects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "lifecycle-image",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"image": "old/image",
						},
					},
				},
				KpConfig: config.KpConfig{
					CanonicalRepository: "gcr.io/my-cool-repo",
					ServiceAccount: corev1.ObjectReference{
						Namespace: "kpack",
						Name:      "some-service-account",
					},
				},
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

				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: annotate(t, &corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "lifecycle-image",
								Namespace: "kpack",
							},
							Data: map[string]string{
								"image": fmt.Sprintf("gcr.io/my-cool-repo/lifecycle@sha256:%s", lifecycleDigest),
							},
						}, timestampAnnotation),
					},
				},
				ExpectCreates: []runtime.Object{
					annotate(t, &v1alpha1.ClusterStore{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStore",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{
								{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", dotnetCoreDigest)},
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "base"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "default"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
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
				K8sObjects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "lifecycle-image",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"image": "old/image",
						},
					},
				},
				KpConfig: config.KpConfig{
					CanonicalRepository: "gcr.io/my-cool-repo",
					ServiceAccount: corev1.ObjectReference{
						Namespace: "kpack",
						Name:      "some-service-account",
					},
				},
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
					annotate(t, &v1alpha1.ClusterStore{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStore",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{
								{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", dotnetCoreDigest)},
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "base"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "default"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
				},
			}.TestImporter(t)
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
				K8sObjects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "lifecycle-image",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"image": "old/image",
						},
					},
				},
				KpackObjects: []runtime.Object{
					annotate(t, &v1alpha1.ClusterStore{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStore",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{
								{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", dotnetCoreDigest)},
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
						Status: v1alpha1.ClusterStackStatus{
							Status: corev1alpha1.Status{},
							ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
								BuildImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									Image:       "",
								},
								RunImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									Image:       "",
								},
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
						Status: v1alpha1.ClusterStackStatus{
							Status: corev1alpha1.Status{},
							ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
								BuildImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									Image:       "",
								},
								RunImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									Image:       "",
								},
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "base"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "default"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
				},
				KpConfig: config.KpConfig{
					CanonicalRepository: "gcr.io/my-cool-repo",
					ServiceAccount: corev1.ObjectReference{
						Namespace: "kpack",
						Name:      "some-service-account",
					},
				},
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

				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: annotate(t, &corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "lifecycle-image",
								Namespace: "kpack",
							},
							Data: map[string]string{
								"image": fmt.Sprintf("gcr.io/my-cool-repo/lifecycle@sha256:%s", newLifecycleDigest),
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterStore{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterStore",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "default",
							},
							Spec: v1alpha1.ClusterStoreSpec{
								Sources: []v1alpha1.StoreImage{
									{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", dotnetCoreDigest)},
									{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", newDotnetCoreDigest)},
									{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "node_js", nodejsDigest)},
								},
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterStack{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterStack",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "base",
							},
							Spec: v1alpha1.ClusterStackSpec{
								Id: stackId,
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", newBuildImageDigest),
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", newRunImageDigest),
								},
							},
							Status: v1alpha1.ClusterStackStatus{
								Status: corev1alpha1.Status{},
								ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
									BuildImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									},
									RunImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									},
								},
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterStack{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterStack",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "default",
							},
							Spec: v1alpha1.ClusterStackSpec{
								Id: stackId,
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", newBuildImageDigest),
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", newRunImageDigest),
								},
							},
							Status: v1alpha1.ClusterStackStatus{
								Status: corev1alpha1.Status{},
								ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
									BuildImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									},
									RunImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									},
								},
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterBuilder{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterBuilder",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "base",
							},
							Spec: v1alpha1.ClusterBuilderSpec{
								BuilderSpec: v1alpha1.BuilderSpec{
									Tag: path.Join("gcr.io/my-cool-repo", "base"),
									Stack: corev1.ObjectReference{
										Kind: "ClusterStack",
										Name: "base",
									},
									Store: corev1.ObjectReference{
										Kind: "ClusterStore",
										Name: "default",
									},
									Order: []v1alpha1.OrderEntry{
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/dotnet-core",
													},
													Optional: false,
												},
											},
										},
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/nodejs",
													},
													Optional: false,
												},
											},
										},
									},
								},
								ServiceAccountRef: corev1.ObjectReference{
									Namespace: "kpack",
									Name:      "some-service-account",
								},
							},
						}, kubectlAnnotation, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterBuilder{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterBuilder",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "default",
							},
							Spec: v1alpha1.ClusterBuilderSpec{
								BuilderSpec: v1alpha1.BuilderSpec{
									Tag: path.Join("gcr.io/my-cool-repo", "default"),
									Stack: corev1.ObjectReference{
										Kind: "ClusterStack",
										Name: "base",
									},
									Store: corev1.ObjectReference{
										Kind: "ClusterStore",
										Name: "default",
									},
									Order: []v1alpha1.OrderEntry{
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/dotnet-core",
													},
													Optional: false,
												},
											},
										},
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/nodejs",
													},
													Optional: false,
												},
											},
										},
									},
								},
								ServiceAccountRef: corev1.ObjectReference{
									Namespace: "kpack",
									Name:      "some-service-account",
								},
							},
						}, kubectlAnnotation, timestampAnnotation),
					},
				},
			}.TestImporter(t)
		})

		it("can import v1alpha1 on an existing cluster", func() {
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
				K8sObjects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "lifecycle-image",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"image": "old/image",
						},
					},
				},
				KpackObjects: []runtime.Object{
					annotate(t, &v1alpha1.ClusterStore{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStore",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{
								{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", dotnetCoreDigest)},
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
						Status: v1alpha1.ClusterStackStatus{
							Status: corev1alpha1.Status{},
							ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
								BuildImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									Image:       "",
								},
								RunImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									Image:       "",
								},
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterStack{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterStack",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterStackSpec{
							Id: stackId,
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
							},
						},
						Status: v1alpha1.ClusterStackStatus{
							Status: corev1alpha1.Status{},
							ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
								BuildImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									Image:       "",
								},
								RunImage: v1alpha1.ClusterStackStatusImage{
									LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									Image:       "",
								},
							},
						},
					}, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "base",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "base"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
					annotate(t, &v1alpha1.ClusterBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterBuilder",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "default",
						},
						Spec: v1alpha1.ClusterBuilderSpec{
							BuilderSpec: v1alpha1.BuilderSpec{
								Tag: path.Join("gcr.io/my-cool-repo", "default"),
								Stack: corev1.ObjectReference{
									Kind: "ClusterStack",
									Name: "base",
								},
								Store: corev1.ObjectReference{
									Kind: "ClusterStore",
									Name: "default",
								},
								Order: []v1alpha1.OrderEntry{
									{
										[]v1alpha1.BuildpackRef{
											{
												BuildpackInfo: v1alpha1.BuildpackInfo{
													Id: "tanzu-buildpacks/dotnet-core",
												},
												Optional: false,
											},
										},
									},
								},
							},
							ServiceAccountRef: corev1.ObjectReference{
								Namespace: "kpack",
								Name:      "some-service-account",
							},
						},
					}, kubectlAnnotation, timestampAnnotation),
				},
				KpConfig: config.KpConfig{
					CanonicalRepository: "gcr.io/my-cool-repo",
					ServiceAccount: corev1.ObjectReference{
						Namespace: "kpack",
						Name:      "some-service-account",
					},
				},
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
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: annotate(t, &v1alpha1.ClusterStore{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterStore",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "default",
							},
							Spec: v1alpha1.ClusterStoreSpec{
								Sources: []v1alpha1.StoreImage{
									{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", dotnetCoreDigest)},
									{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "dotnet_core", newDotnetCoreDigest)},
									{Image: fmt.Sprintf("gcr.io/my-cool-repo/%s@sha256:%s", "node_js", nodejsDigest)},
								},
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterStack{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterStack",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "base",
							},
							Spec: v1alpha1.ClusterStackSpec{
								Id: stackId,
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", newBuildImageDigest),
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", newRunImageDigest),
								},
							},
							Status: v1alpha1.ClusterStackStatus{
								Status: corev1alpha1.Status{},
								ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
									BuildImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									},
									RunImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									},
								},
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterStack{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterStack",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "default",
							},
							Spec: v1alpha1.ClusterStackSpec{
								Id: stackId,
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", newBuildImageDigest),
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", newRunImageDigest),
								},
							},
							Status: v1alpha1.ClusterStackStatus{
								Status: corev1alpha1.Status{},
								ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
									BuildImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/build@sha256:%s", buildImageDigest),
									},
									RunImage: v1alpha1.ClusterStackStatusImage{
										LatestImage: fmt.Sprintf("gcr.io/my-cool-repo/run@sha256:%s", runImageDigest),
									},
								},
							},
						}, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterBuilder{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterBuilder",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "base",
							},
							Spec: v1alpha1.ClusterBuilderSpec{
								BuilderSpec: v1alpha1.BuilderSpec{
									Tag: path.Join("gcr.io/my-cool-repo", "base"),
									Stack: corev1.ObjectReference{
										Kind: "ClusterStack",
										Name: "base",
									},
									Store: corev1.ObjectReference{
										Kind: "ClusterStore",
										Name: "default",
									},
									Order: []v1alpha1.OrderEntry{
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/dotnet-core",
													},
													Optional: false,
												},
											},
										},
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/nodejs",
													},
													Optional: false,
												},
											},
										},
									},
								},
								ServiceAccountRef: corev1.ObjectReference{
									Namespace: "kpack",
									Name:      "some-service-account",
								},
							},
						}, kubectlAnnotation, timestampAnnotation),
					},
					{
						Object: annotate(t, &v1alpha1.ClusterBuilder{
							TypeMeta: metav1.TypeMeta{
								Kind:       "ClusterBuilder",
								APIVersion: "kpack.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name: "default",
							},
							Spec: v1alpha1.ClusterBuilderSpec{
								BuilderSpec: v1alpha1.BuilderSpec{
									Tag: path.Join("gcr.io/my-cool-repo", "default"),
									Stack: corev1.ObjectReference{
										Kind: "ClusterStack",
										Name: "base",
									},
									Store: corev1.ObjectReference{
										Kind: "ClusterStore",
										Name: "default",
									},
									Order: []v1alpha1.OrderEntry{
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/dotnet-core",
													},
													Optional: false,
												},
											},
										},
										{
											[]v1alpha1.BuildpackRef{
												{
													BuildpackInfo: v1alpha1.BuildpackInfo{
														Id: "tanzu-buildpacks/nodejs",
													},
													Optional: false,
												},
											},
										},
									},
								},
								ServiceAccountRef: corev1.ObjectReference{
									Namespace: "kpack",
									Name:      "some-service-account",
								},
							},
						}, kubectlAnnotation, timestampAnnotation),
					},
				},
			}.TestImporter(t)
		})

		it("does not create any resources if any relocation fails", func() {
			TestImport{
				Images: map[string]v1.Image{
					"new-image.com/lifecycle":              fakes.NewFakeImage(lifecycleDigest),
					"new-image.com/buildpacks/dotnet-core": fakes.NewFakeLabeledImage("io.buildpacks.buildpackage.metadata", fmt.Sprintf("{\"id\":%q}", dotnetCoreId), dotnetCoreDigest),
					"new-image.com/stacks/base/build":      fakes.NewFakeLabeledImage("io.buildpacks.stack.id", stackId, buildImageDigest),
				},
				K8sObjects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "lifecycle-image",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"image": "old/image",
						},
					},
				},
				KpConfig: config.KpConfig{
					CanonicalRepository: "gcr.io/my-cool-repo",
					ServiceAccount: corev1.ObjectReference{
						Namespace: "kpack",
						Name:      "some-service-account",
					},
				},
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

				ExpectUpdates: []clientgotesting.UpdateActionImpl{},
				ExpectCreates: []runtime.Object{},
				ExpectErr:     errors.New("buddy we don't have your image, check another registry"),
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
				K8sObjects: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "lifecycle-image",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"image": "old/image",
						},
					},
				},
				KpConfig: config.KpConfig{
					CanonicalRepository: "gcr.io/my-cool-repo",
					ServiceAccount: corev1.ObjectReference{
						Namespace: "kpack",
						Name:      "some-service-account",
					},
				},
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

				ExpectUpdates: []clientgotesting.UpdateActionImpl{},
				ExpectCreates: []runtime.Object{},
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
	K8sObjects           []runtime.Object
	KpackObjects         []runtime.Object
	KpConfig             config.KpConfig
	DependencyDescriptor string
	DryRun               bool
	ExpectUpdates        []clientgotesting.UpdateActionImpl
	Images               map[string]v1.Image
	ExpectCreates        []runtime.Object
	ExpectErr            error
}

func (i TestImport) TestImporter(t *testing.T) {
	t.Helper()
	client := kpackfakes.NewSimpleClientset(i.KpackObjects...)
	k8sClient := k8sfakes.NewSimpleClientset(i.K8sObjects...)

	buffer := &bytes.Buffer{}
	var err error
	importer := NewImporter(testLogger{writer: buffer}, k8sClient, client, &fakeFetcher{Images: i.Images}, &fakeRelocator{}, &fakeWaiter{}, &fakeTimestampProvider{ts: time.Time{}.String()})
	if i.DryRun {
		_, err = importer.ImportDescriptorDryRun(context.Background(), authn.NewMultiKeychain(), i.KpConfig, strings.NewReader(i.DependencyDescriptor))
	} else {
		_, err = importer.ImportDescriptor(context.Background(), authn.NewMultiKeychain(), i.KpConfig, strings.NewReader(i.DependencyDescriptor))
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
		nil,
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
