package commands

import (
	"errors"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/kmeta"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

func TestWaiter(t *testing.T) {
	spec.Run(t, "Waiter", testWaiter)
}

func init() {
	v1alpha1.AddToScheme(scheme.Scheme)
}

func testWaiter(t *testing.T, when spec.G, it spec.S) {
	var (
		watcher       *TestWatcher
		generation    int64 = 2
		dynamicClient       = dynamicfake.NewSimpleDynamicClient(scheme.Scheme)
		waiter              = NewWaiter(dynamicClient)
	)

	when("Wait", func() {
		var resourceToWatch *v1alpha1.Builder

		it.Before(func() {
			resourceToWatch = &v1alpha1.Builder{
				TypeMeta: v1.TypeMeta{
					Kind: "Builder",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:            "some-name",
					Namespace:       "some-namespace",
					ResourceVersion: "1",
					Generation:      generation,
				},
			}
			watcher = &TestWatcher{
				events:           make(chan watch.Event, 100),
				expectedResource: resourceToWatch,
			}
			dynamicClient.PrependWatchReactor("builders", watcher.watchReactor)
		})

		it("returns no error when resource is already ready", func() {
			resourceToWatch.Status = v1alpha1.BuilderStatus{
				Status: conditionReady(corev1.ConditionTrue, generation),
			}

			assert.NoError(t, waiter.Wait(resourceToWatch))
		})

		it("returns an error when resource is already failed", func() {
			resourceToWatch.Status = v1alpha1.BuilderStatus{
				Status: conditionReady(corev1.ConditionFalse, generation),
			}

			assert.EqualError(t, waiter.Wait(resourceToWatch), "Builder \"some-name\" not ready: some-message")
		})

		it("waits for the correct generation", func() {
			resourceToWatch.Status = v1alpha1.BuilderStatus{
				Status: conditionReady(corev1.ConditionFalse, generation-1),
			}

			watcher.addEvent(watch.Event{
				Type: watch.Modified,
				Object: &v1alpha1.Builder{
					TypeMeta:   resourceToWatch.TypeMeta,
					ObjectMeta: resourceToWatch.ObjectMeta,
					Status:     v1alpha1.BuilderStatus{Status: conditionReady(corev1.ConditionTrue, generation)},
				},
			})

			assert.NoError(t, waiter.Wait(resourceToWatch))
		})
	})

	when("BuilderWait", func() {
		var (
			cbToWatch       *v1alpha1.ClusterBuilder
			storeGeneration int64 = 2
			stackGeneration int64 = 2
		)

		it.Before(func() {
			cbToWatch = &v1alpha1.ClusterBuilder{
				TypeMeta: v1.TypeMeta{
					Kind: "ClusterBuilder",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:            "some-name",
					ResourceVersion: "1",
				},
			}
			watcher = &TestWatcher{
				events:           make(chan watch.Event, 100),
				expectedResource: cbToWatch,
			}
			dynamicClient.PrependWatchReactor("clusterbuilders", watcher.watchReactor)
		})

		it("returns no error when resource is ready", func() {
			cbToWatch.Status = v1alpha1.BuilderStatus{
				ObservedStoreGeneration: storeGeneration,
				ObservedStackGeneration: stackGeneration,
				Status:                  conditionReady(corev1.ConditionTrue, generation),
			}

			assert.NoError(t, waiter.BuilderWait(cbToWatch, storeGeneration, stackGeneration))
		})

		it("returns an error when resource has failed", func() {
			cbToWatch.Status = v1alpha1.BuilderStatus{
				ObservedStoreGeneration: storeGeneration,
				ObservedStackGeneration: stackGeneration,
				Status:                  conditionReady(corev1.ConditionFalse, generation),
			}

			assert.EqualError(t, waiter.BuilderWait(cbToWatch, storeGeneration, stackGeneration), "ClusterBuilder \"some-name\" not ready: some-message")
		})

		it("waits for the correct generations", func() {
			cbToWatch.Status = v1alpha1.BuilderStatus{
				ObservedStoreGeneration: storeGeneration - 1,
				ObservedStackGeneration: stackGeneration - 1,
				Status:                  conditionReady(corev1.ConditionFalse, generation-1),
			}

			watcher.addEvent(watch.Event{
				Type: watch.Modified,
				Object: &v1alpha1.Builder{
					TypeMeta:   cbToWatch.TypeMeta,
					ObjectMeta: cbToWatch.ObjectMeta,
					Status: v1alpha1.BuilderStatus{
						ObservedStoreGeneration: storeGeneration,
						ObservedStackGeneration: stackGeneration,
						Status:                  conditionReady(corev1.ConditionTrue, generation),
					},
				},
			})

			assert.NoError(t, waiter.BuilderWait(cbToWatch, storeGeneration, stackGeneration))
		})

		it("returns no error when observedStack/Store generation is 0 (is not supported)", func() {
			cbToWatch.Status = v1alpha1.BuilderStatus{
				ObservedStoreGeneration: 0,
				ObservedStackGeneration: 0,
				Status:                  conditionReady(corev1.ConditionTrue, generation),
			}

			assert.NoError(t, waiter.BuilderWait(cbToWatch, storeGeneration, stackGeneration))
		})
	})
}

func conditionReady(status corev1.ConditionStatus, generation int64) corev1alpha1.Status {
	return corev1alpha1.Status{
		ObservedGeneration: generation,
		Conditions: []corev1alpha1.Condition{
			{
				Type:    corev1alpha1.ConditionReady,
				Status:  status,
				Message: "some-message",
			},
		},
	}
}

type TestWatcher struct {
	events           chan watch.Event
	expectedResource kmeta.OwnerRefable
}

func (t *TestWatcher) addEvent(event watch.Event) {
	t.events <- event
}

func (t *TestWatcher) Stop() {
}

func (t *TestWatcher) ResultChan() <-chan watch.Event {
	return t.events
}

func (t *TestWatcher) watchReactor(action clientgotesting.Action) (handled bool, ret watch.Interface, err error) {
	if t.expectedResource == nil {
		return false, nil, errors.New("test watcher must be configured with an expected resource to be used")
	}

	watchAction := action.(clientgotesting.WatchAction)
	if watchAction.GetWatchRestrictions().ResourceVersion != t.expectedResource.GetObjectMeta().GetResourceVersion() {
		return false, nil, errors.New("expected watch on resource version")
	}

	match, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
	if !found {
		return false, nil, errors.New("expected watch on name")
	}
	if match != t.expectedResource.GetObjectMeta().GetName() {
		return false, nil, errors.New("expected watch on name")
	}

	return true, t, nil
}
