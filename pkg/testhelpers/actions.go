package testhelpers

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
)

type Actions struct {
	Gets              []clientgotesting.GetAction
	Creates           []clientgotesting.CreateAction
	Updates           []clientgotesting.UpdateAction
	Deletes           []clientgotesting.DeleteAction
	DeleteCollections []clientgotesting.DeleteCollectionAction
	Patches           []clientgotesting.PatchAction
}

// ActionRecorder contains list of K8s request actions.
type ActionRecorder interface {
	Actions() []clientgotesting.Action
}

// ActionRecorderList is a list of ActionRecorder objects.
type ActionRecorderList []ActionRecorder

// ActionsByVerb fills in Actions objects, sorting the actions
// by verb.
func (l ActionRecorderList) ActionsByVerb() (Actions, error) {
	var a Actions

	for _, recorder := range l {
		for _, action := range recorder.Actions() {
			switch action.GetVerb() {
			case "get":
				if get, ok := action.(clientgotesting.GetAction); ok {
					a.Gets = append(a.Gets, get)
				}
			case "create":
				a.Creates = append(a.Creates,
					action.(clientgotesting.CreateAction))
			case "update":
				a.Updates = append(a.Updates,
					action.(clientgotesting.UpdateAction))
			case "delete":
				a.Deletes = append(a.Deletes,
					action.(clientgotesting.DeleteAction))
			case "delete-collection":
				a.DeleteCollections = append(a.DeleteCollections,
					action.(clientgotesting.DeleteCollectionAction))
			case "patch":
				a.Patches = append(a.Patches,
					action.(clientgotesting.PatchAction))
			case "list", "watch": // avoid 'unexpected verb list/watch' error
			default:
				return a, fmt.Errorf("unexpected verb %v: %+v", action.GetVerb(), action)
			}
		}
	}
	return a, nil
}

func TestK8sActions(t *testing.T, clientset *k8sfakes.Clientset, expectUpdates []clientgotesting.UpdateActionImpl, expectCreates []runtime.Object, expectDeletes []clientgotesting.DeleteActionImpl) {
	t.Helper()
	actions, err := ActionRecorderList{clientset}.ActionsByVerb()
	require.NoError(t, err)

	for i, want := range expectCreates {
		if i >= len(actions.Creates) {
			t.Errorf("Missing create: %#v", want)
			continue
		}

		got := actions.Creates[i].GetObject()

		if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected create (-want, +got): %s", diff)
		}
	}

	if got, want := len(actions.Creates), len(expectCreates); got > want {
		for _, extra := range actions.Creates[want:] {
			t.Errorf("Extra create: %#v", extra.GetObject())
		}
	}

	for i, want := range expectDeletes {
		if i >= len(actions.Deletes) {
			wo := want.GetName()
			t.Errorf("Missing delete for %#v", wo)
			continue
		}

		gotNamespace := actions.Deletes[i].GetNamespace()
		if diff := cmp.Diff(want.GetNamespace(), gotNamespace, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected delete (-want, +got): %s", diff)
		}

		gotName := actions.Deletes[i].GetName()
		if diff := cmp.Diff(want.GetName(), gotName, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected delete (-want, +got): %s", diff)
		}
	}

	if got, want := len(actions.Deletes), len(expectDeletes); got > want {
		for _, extra := range actions.Deletes[want:] {
			t.Errorf("Extra delete: %s/%s", extra.GetNamespace(), extra.GetName())
		}
	}

	for i, want := range expectUpdates {
		if i >= len(actions.Updates) {
			wo := want.GetObject()
			t.Errorf("Missing update for %#v", wo)
			continue
		}

		got := actions.Updates[i].GetObject()

		if diff := cmp.Diff(want.GetObject(), got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected update (-want, +got): %s", diff)
		}
	}

	if got, want := len(actions.Updates), len(expectUpdates); got > want {
		for _, extra := range actions.Updates[want:] {
			t.Errorf("Extra update: %#v", extra.GetObject())
		}
	}

	for _, extra := range actions.DeleteCollections {
		t.Errorf("Extra delete-collection: %#v", extra)
	}

	for _, extra := range actions.Patches {
		t.Errorf("Extra patch: %#v; raw: %s", extra, string(extra.GetPatch()))
	}
}

func TestKpackActions(t *testing.T, clientset *kpackfakes.Clientset, expectUpdates []clientgotesting.UpdateActionImpl, expectCreates []runtime.Object, expectDeletes []clientgotesting.DeleteActionImpl) {
	t.Helper()
	actions, err := ActionRecorderList{clientset}.ActionsByVerb()
	require.NoError(t, err)

	for i, want := range expectCreates {
		if i >= len(actions.Creates) {
			t.Errorf("Missing create: %#v", want)
			continue
		}

		got := actions.Creates[i].GetObject()

		if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected create (-want, +got): %s", diff)
		}
	}

	if got, want := len(actions.Creates), len(expectCreates); got > want {
		for _, extra := range actions.Creates[want:] {
			t.Errorf("Extra create: %#v", extra.GetObject())
		}
	}

	for i, want := range expectDeletes {
		if i >= len(actions.Deletes) {
			wo := want.GetName()
			t.Errorf("Missing delete for %#v", wo)
			continue
		}

		gotNamespace := actions.Deletes[i].GetNamespace()
		if diff := cmp.Diff(want.GetNamespace(), gotNamespace, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected delete (-want, +got): %s", diff)
		}

		gotName := actions.Deletes[i].GetName()
		if diff := cmp.Diff(want.GetName(), gotName, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected delete (-want, +got): %s", diff)
		}
	}

	if got, want := len(actions.Deletes), len(expectDeletes); got > want {
		for _, extra := range actions.Deletes[want:] {
			t.Errorf("Extra delete: %s/%s", extra.GetNamespace(), extra.GetName())
		}
	}

	for i, want := range expectUpdates {
		if i >= len(actions.Updates) {
			wo := want.GetObject()
			t.Errorf("Missing update for %#v", wo)
			continue
		}

		got := actions.Updates[i].GetObject()

		if diff := cmp.Diff(want.GetObject(), got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected update (-want, +got): %s", diff)
		}
	}

	if got, want := len(actions.Updates), len(expectUpdates); got > want {
		for _, extra := range actions.Updates[want:] {
			t.Errorf("Extra update: %#v", extra.GetObject())
		}
	}

	for _, extra := range actions.DeleteCollections {
		t.Errorf("Extra delete-collection: %#v", extra)
	}

	for _, extra := range actions.Patches {
		t.Errorf("Extra patch: %#v; raw: %s", extra, string(extra.GetPatch()))
	}
}
