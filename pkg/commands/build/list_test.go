package build_test

import (
	"testing"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/build"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuildListCommand(t *testing.T) {
	spec.Run(t, "TestBuildListCommand", testBuildListCommand)
}

func testBuildListCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		image            = "test-image"
		defaultNamespace = "some-default-namespace"
		expectedOutput   = `BUILD    STATUS      IMAGE                   STARTED                FINISHED               REASON
1        SUCCESS     repo.com/image-1:tag    0001-01-01 00:00:00    0001-01-01 00:00:00    CONFIG
2        FAILURE     repo.com/image-2:tag    0001-01-01 01:00:00    0001-01-01 00:00:00    COMMIT+
3        BUILDING    repo.com/image-3:tag    0001-01-01 05:00:00                           TRIGGER
`
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return build.NewListCommand(clientSet, defaultNamespace)
	}

	when("listing builds", func() {
		when("in the default namespace", func() {
			when("there are builds", func() {
				it("lists the builds", func() {
					testhelpers.CommandTest{
						Objects:        makeTestBuilds(image, defaultNamespace),
						Args:           []string{image},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builds", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{image},
						ExpectErr:      true,
						ExpectedOutput: "Error: no builds for image \"test-image\" found in \"some-default-namespace\" namespace\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("in a given namespace", func() {
			const namespace = "some-namespace"

			when("there are builds", func() {
				it("lists the builds", func() {
					testhelpers.CommandTest{
						Objects:        makeTestBuilds(image, namespace),
						Args:           []string{image, "-n", namespace},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builds", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{image, "-n", namespace},
						ExpectErr:      true,
						ExpectedOutput: "Error: no builds for image \"test-image\" found in \"some-namespace\" namespace\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

func makeTestBuilds(image string, namespace string) []runtime.Object {
	buildOne := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-one",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.ImageLabel:       image,
				v1alpha1.BuildNumberLabel: "1",
			},
			Annotations: map[string]string{
				v1alpha1.BuildReasonAnnotation: "CONFIG",
			},
		},
		Status: v1alpha1.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionTrue,
						LastTransitionTime: corev1alpha1.VolatileTime{
							Inner: metav1.Time{},
						},
					},
				},
			},
			LatestImage: "repo.com/image-1:tag",
		},
	}
	buildTwo := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "build-two",
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Time{}.Add(1 * time.Hour)},
			Labels: map[string]string{
				v1alpha1.ImageLabel:       image,
				v1alpha1.BuildNumberLabel: "2",
			},
			Annotations: map[string]string{
				v1alpha1.BuildReasonAnnotation: "COMMIT,BUILDPACK",
			},
		},
		Status: v1alpha1.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionFalse,
						LastTransitionTime: corev1alpha1.VolatileTime{
							Inner: metav1.Time{},
						},
					},
				},
			},
			LatestImage: "repo.com/image-2:tag",
		},
	}
	buildThree := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "build-three",
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
			Labels: map[string]string{
				v1alpha1.ImageLabel:       image,
				v1alpha1.BuildNumberLabel: "3",
			},
			Annotations: map[string]string{
				v1alpha1.BuildReasonAnnotation: "TRIGGER",
			},
		},
		Status: v1alpha1.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: corev1alpha1.Conditions{
					{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
			LatestImage: "repo.com/image-3:tag",
		},
	}
	ignoredBuild := &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ignored",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.ImageLabel: "some-other-image",
			},
		},
	}
	return []runtime.Object{buildOne, buildThree, buildTwo, ignoredBuild}
}
