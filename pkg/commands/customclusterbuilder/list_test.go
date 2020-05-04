package customclusterbuilder_test

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

	"github.com/pivotal/build-service-cli/pkg/commands/customclusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderListCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderListCommand", testClusterBuilderListCommand)
}

func testClusterBuilderListCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		expectedOutput = `NAME              READY    STACK                          IMAGE
test-builder-1    true     io.buildpacks.stacks.centos    some-registry.com/test-builder-1:tag
test-builder-2    false                                   
test-builder-3    true     io.buildpacks.stacks.bionic    some-registry.com/test-builder-3:tag

`
	)

	var (
		customClusterBuilder1 = &expv1alpha1.CustomClusterBuilder{
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
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
						ID:       "io.buildpacks.stacks.centos",
					},
					LatestImage: "some-registry.com/test-builder-1:tag",
				},
			},
		}
		customClusterBuilder2 = &expv1alpha1.CustomClusterBuilder{
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
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
				},
			},
		}
		customClusterBuilder3 = &expv1alpha1.CustomClusterBuilder{
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
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9fasdfa847098234",
						ID:       "io.buildpacks.stacks.bionic",
					},
					LatestImage: "some-registry.com/test-builder-3:tag",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return customclusterbuilder.NewListCommand(clientSet)
	}

	when("listing clusterbuilder", func() {
		when("there are clusterbuilders", func() {
			it("lists the builders", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						customClusterBuilder1,
						customClusterBuilder2,
						customClusterBuilder3,
					},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("there are no clusterbuilders", func() {
			it("prints an appropriate message", func() {
				testhelpers.CommandTest{
					ExpectErr:      true,
					ExpectedOutput: "Error: no clusterbuilders found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}
