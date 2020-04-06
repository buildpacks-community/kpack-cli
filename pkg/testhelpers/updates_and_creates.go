package testhelpers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

func TestUpdatesAndCreates(t *testing.T, clientset *fake.Clientset, expectUpdates []clientgotesting.UpdateActionImpl, expectCreates []runtime.Object) {
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

	for _, extra := range actions.Deletes {
		t.Errorf("Extra delete: %s/%s", extra.GetNamespace(), extra.GetName())
	}

	for _, extra := range actions.DeleteCollections {
		t.Errorf("Extra delete-collection: %#v", extra)
	}

	for _, extra := range actions.Patches {
		t.Errorf("Extra patch: %#v; raw: %s", extra, string(extra.GetPatch()))
	}
}
