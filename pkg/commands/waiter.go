package commands

import (
	"context"
	"fmt"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchTools "k8s.io/client-go/tools/watch"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

type ResourceWaiter interface {
	Wait(runtime.Object) error
	BuilderWait(runtime.Object, int64, int64) error
}

func NewResourceWaiter(dc dynamic.Interface) ResourceWaiter {
	return NewWaiter(dc)
}

type Waiter struct {
	dynamicClient dynamic.Interface
}

func NewWaiter(dc dynamic.Interface) *Waiter {
	return &Waiter{dynamicClient: dc}
}

func (i *Waiter) Wait(ob runtime.Object) error {
	m, ok := ob.(kmeta.OwnerRefable)
	if !ok {
		return errors.New("unexpected type")
	}
	return i.wait(ob, hasResolved(m.GetObjectMeta().GetGeneration()))
}

// Used to ensure the builder has resolved to the expected stack and store (importing)
func (i *Waiter) BuilderWait(ob runtime.Object, expectedStoreGeneration, expectedStackGeneration int64) error {
	m, ok := ob.(kmeta.OwnerRefable)
	if !ok {
		return errors.New("unexpected type")
	}
	return i.wait(ob, builderHasResolved(m.GetObjectMeta().GetGeneration(), expectedStoreGeneration, expectedStackGeneration))
}

func (i *Waiter) wait(ob runtime.Object, cf watchTools.ConditionFunc) error {
	e := &watch.Event{Object: ob}
	done, err := cf(*e)
	if err != nil {
		return err
	}
	if !done {
		m, ok := ob.(kmeta.OwnerRefable)
		if !ok {
			return errors.New("unexpected type")
		}
		name := m.GetObjectMeta().GetName()
		gvr, _ := meta.UnsafeGuessKindToResource(m.GetGroupVersionKind())
		rv := m.GetObjectMeta().GetResourceVersion()
		w := watchOne{name: name, gvr: gvr, dynamicClient: i.dynamicClient}
		e, err = watchTools.Until(context.Background(), rv, w, filterErrors(cf))
		if err != nil {
			return err
		}
	}

	kr, err := eventToDuck(e)
	if err != nil {
		return err
	}

	if condition := kr.Status.GetCondition(apis.ConditionReady); condition.IsFalse() {
		if condition.Message != "" {
			return errors.Errorf("%v %q not ready: %v", kr.Kind, kr.Name, condition.Message)
		}

		return errors.Errorf("%v %q not ready", kr.Kind, kr.Name)
	}
	return nil
}

func eventToDuck(e *watch.Event) (*duckv1.KResource, error) {
	u := &unstructured.Unstructured{}
	var err error
	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(e.Object)
	if err != nil {
		return nil, err
	}

	kr := &duckv1.KResource{}
	if err := duck.FromUnstructured(u, kr); err != nil {
		return nil, err
	}
	return kr, nil
}

func hasResolved(generation int64) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		kr, err := eventToDuck(&event)
		if err != nil {
			return false, err
		}

		if kr.Status.ObservedGeneration < generation {
			return false, nil // still waiting on update
		}

		if kr.Status.GetCondition(apis.ConditionReady).IsUnknown() {
			return false, nil
		}

		return true, nil
	}
}

type builderWaitable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status v1alpha1.BuilderStatus `json:"status"`
}

func builderHasResolved(builderGeneration, expectedStoreGen, expectedStackGen int64) func(event watch.Event) (bool, error) {
	return func(e watch.Event) (bool, error) {
		u := &unstructured.Unstructured{}
		var err error
		u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(e.Object)
		if err != nil {
			return false, err
		}

		b := builderWaitable{}
		if err := duck.FromUnstructured(u, &b); err != nil {
			return false, err
		}


		if b.Status.ObservedGeneration < builderGeneration ||
			(b.Status.ObservedStoreGeneration != 0 && b.Status.ObservedStoreGeneration < expectedStoreGen) || // ObservedStoreGeneration is 0 when kpack does not support it
			(b.Status.ObservedStackGeneration != 0 && b.Status.ObservedStackGeneration < expectedStackGen) {  // ObservedStackGeneration is 0 when kpack does not support it
			return false, nil // still waiting on update
		}

		if b.Status.GetCondition(corev1alpha1.ConditionReady).IsUnknown() {
			return false, nil
		}

		return true, nil
	}
}

type watchOne struct {
	name          string
	namespace     string
	gvr           schema.GroupVersionResource
	dynamicClient dynamic.Interface
}

func (w watchOne) Watch(options metav1.ListOptions) (watch.Interface, error) {
	options.FieldSelector = fmt.Sprintf("metadata.name=%s", w.name)
	return w.dynamicClient.Resource(w.gvr).Namespace(w.namespace).Watch(options)
}

func filterErrors(condition watchTools.ConditionFunc) watchTools.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Error {
			return false, errors.Errorf("error on watch %+v", event.Object)
		}

		return condition(event)
	}
}
