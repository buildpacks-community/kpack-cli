package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderStatusCommand", testClusterBuilderStatusCommand)
}

func testClusterBuilderStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		expectedReadyOutput = `Status:       Ready
Image:        some-registry.com/test-builder-1:tag
Stack:        io.buildpacks.stacks.centos
Run Image:    gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234

BUILDPACK ID               VERSION
org.cloudfoundry.nodejs    v0.2.1
org.cloudfoundry.go        v0.0.3


DETECTION ORDER              
Group #1                     
  org.cloudfoundry.nodejs    
Group #2                     
  org.cloudfoundry.go        

`
		expectedNotReadyOutput = `Status:    Not Ready
Reason:    this builder is not ready for the purpose of a test

`
		expectedUnkownOutput = `Status:    Unknown

`
	)

	var (
		readyClusterBuilder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomClusterBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder-1",
			},
			Spec: expv1alpha1.CustomClusterBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag:   "some-registry.com/test-builder-1",
					Stack: "test-stack",
					Store: "test-store",
					Order: []expv1alpha1.OrderEntry{
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountRef: corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-service-account",
				},
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					BuilderMetadata: v1alpha1.BuildpackMetadataList{
						{
							Id:      "org.cloudfoundry.nodejs",
							Version: "v0.2.1",
						},
						{
							Id:      "org.cloudfoundry.go",
							Version: "v0.0.3",
						},
					},
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
						ID:       "io.buildpacks.stacks.centos",
					},
					LatestImage: "some-registry.com/test-builder-1:tag",
				},
			},
		}
		notReadyClusterBuilder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomClusterBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder-2",
			},
			Spec: expv1alpha1.CustomClusterBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag:   "some-registry.com/test-builder-2",
					Stack: "test-stack",
					Store: "test-store",
					Order: []expv1alpha1.OrderEntry{
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountRef: corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-service-account",
				},
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:    corev1alpha1.ConditionReady,
								Status:  corev1.ConditionFalse,
								Message: "this builder is not ready for the purpose of a test",
							},
						},
					},
				},
			},
		}
		unknownClusterBuilder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomClusterBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder-3",
			},
			Spec: expv1alpha1.CustomClusterBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag:   "some-registry.com/test-builder-3",
					Stack: "test-stack",
					Store: "test-store",
					Order: []expv1alpha1.OrderEntry{
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountRef: corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-service-account",
				},
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return clusterbuilder.NewStatusCommand(clientSet)
	}

	when("getting clusterbuilder status", func() {
		when("the clusterbuilder exists", func() {
			when("the builder is ready", func() {
				it("shows the build status", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{readyClusterBuilder},
						Args:           []string{"test-builder-1"},
						ExpectedOutput: expectedReadyOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("the builder is not ready", func() {
				it("shows the build status of not ready builder", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{notReadyClusterBuilder},
						Args:           []string{"test-builder-2"},
						ExpectedOutput: expectedNotReadyOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("the builder is unknown", func() {
				it("shows the build status of unknown builder", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{unknownClusterBuilder},
						Args:           []string{"test-builder-3"},
						ExpectedOutput: expectedUnkownOutput,
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("the clusterbuidler does not exist", func() {
			it("prints an appropriate message", func() {
				testhelpers.CommandTest{
					Args:           []string{"non-existant-builder"},
					ExpectErr:      true,
					ExpectedOutput: "Error: customclusterbuilders.experimental.kpack.pivotal.io \"non-existant-builder\" not found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}
