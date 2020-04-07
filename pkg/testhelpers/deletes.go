package testhelpers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	clientgotesting "k8s.io/client-go/testing"
)

func TestDeletes(t *testing.T, clientset *fake.Clientset, expectDeletes []clientgotesting.DeleteActionImpl) {
	t.Helper()
	actions, err := ActionRecorderList{clientset}.ActionsByVerb()
	require.NoError(t, err)

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
}
