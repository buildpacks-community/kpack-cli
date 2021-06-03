// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"fmt"
	"time"

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

const defaultWaitTimeout = 10 * time.Minute

type ResourceWaiter interface {
	Wait(ctx context.Context, object runtime.Object, extraChecks ...watchTools.ConditionFunc) error
}

func NewResourceWaiter(dc dynamic.Interface) ResourceWaiter {
	return NewWaiter(dc, defaultWaitTimeout)
}

type Waiter struct {
	dynamicClient dynamic.Interface
	timeout       time.Duration
	ctx           context.Context
}

func NewWaiter(dc dynamic.Interface, timeout time.Duration) *Waiter {
	return &Waiter{dynamicClient: dc, timeout: timeout}
}

func (w *Waiter) Wait(ctx context.Context, ob runtime.Object, extraConditions ...watchTools.ConditionFunc) error {
	m, ok := ob.(kmeta.OwnerRefable)
	if !ok {
		return errors.New("unexpected type")
	}
	return w.wait(ctx, ob, hasResolved(m.GetObjectMeta().GetGeneration()), extraConditions...)
}

func (w *Waiter) wait(ctx context.Context, ob runtime.Object, condition watchTools.ConditionFunc, extraConditions ...watchTools.ConditionFunc) error {
	e := &watch.Event{Object: ob}
	cfs := append([]watchTools.ConditionFunc{condition}, extraConditions...)
	done, err := runChecks(*e, cfs)
	if err != nil {
		return err
	}

	if !done {
		refable, ok := ob.(kmeta.OwnerRefable)
		if !ok {
			return errors.New("unexpected type")
		}

		rv := refable.GetObjectMeta().GetResourceVersion()
		watchOne := newWatchOneWatcher(ctx, refable, w.dynamicClient)

		ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
		defer cancel()
		e, err = watchTools.Until(ctx, rv, watchOne, filterErrors(cfs)...)
		if err != nil {
			return err
		}
	}

	conditionCheckable, err := eventToDuck(e)
	if err != nil {
		return err
	}

	if cond := conditionCheckable.Status.GetCondition(apis.ConditionReady); cond.IsFalse() {
		if cond.Message != "" {
			return errors.Errorf("%v %q not ready: %v", conditionCheckable.Kind, conditionCheckable.Name, cond.Message)
		}

		return errors.Errorf("%v %q not ready", conditionCheckable.Kind, conditionCheckable.Name)
	}
	return nil
}

func runChecks(e watch.Event, cfs []watchTools.ConditionFunc) (bool, error) {
	for _, cf := range cfs {
		done, err := cf(e)
		if err != nil {
			return false, err
		}
		if !done {
			return false, nil
		}
	}
	return true, nil
}

func eventToDuck(e *watch.Event) (*duckv1.KResource, error) {
	u := &unstructured.Unstructured{}
	var err error
	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(e.Object)
	if err != nil {
		return nil, err
	}

	statusDuck := &duckv1.KResource{}
	if err := duck.FromUnstructured(u, statusDuck); err != nil {
		return nil, err
	}
	return statusDuck, nil
}

func hasResolved(expectedGeneration int64) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		genObservable, err := eventToDuck(&event)
		if err != nil {
			return false, err
		}

		if genObservable.Status.ObservedGeneration < expectedGeneration {
			return false, nil // still waiting on update
		}

		if genObservable.Status.GetCondition(apis.ConditionReady).IsUnknown() {
			return false, nil
		}

		return true, nil
	}
}

type watchOneWatcher struct {
	name          string
	namespace     string
	gvr           schema.GroupVersionResource
	dynamicClient dynamic.Interface
	ctx           context.Context
}

func newWatchOneWatcher(ctx context.Context, object kmeta.OwnerRefable, client dynamic.Interface) watchOneWatcher {
	name := object.GetObjectMeta().GetName()
	namespace := object.GetObjectMeta().GetNamespace()
	gvr, _ := meta.UnsafeGuessKindToResource(object.GetGroupVersionKind())
	return watchOneWatcher{ctx: ctx, name: name, namespace: namespace, gvr: gvr, dynamicClient: client}
}

func (w watchOneWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	options.FieldSelector = fmt.Sprintf("metadata.name=%s", w.name)
	return w.dynamicClient.Resource(w.gvr).Namespace(w.namespace).Watch(w.ctx, options)
}

func filterErrors(conditions []watchTools.ConditionFunc) []watchTools.ConditionFunc {
	cfs := []watchTools.ConditionFunc{}
	for _, c := range conditions {
		cfs = append(cfs, func(event watch.Event) (bool, error) {
			if event.Type == watch.Error {
				return false, errors.Errorf("error on watch %+v", event.Object)
			}

			return c(event)
		})
	}
	return cfs
}

type noopWaiter struct{}

func NewNoopWaiter() ResourceWaiter {
	return &noopWaiter{}
}

func (n noopWaiter) Wait(ctx context.Context, object runtime.Object, extraChecks ...watchTools.ConditionFunc) error {
	return nil
}
