package config

import (
	"context"
	"testing"

	kpacktesthelpers "github.com/pivotal/kpack/pkg/reconciler/testhelpers"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
)

func TestKpConfigProvider(t *testing.T) {
	spec.Run(t, "TestKpConfigProvider", testKpConfigProvider)
}

func testKpConfigProvider(t *testing.T, when spec.G, it spec.S) {
	it("uses the new keys in the existing kp config if they exist", func() {
		kpConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"default.repository":                          "some-repo",
				"default.repository.serviceaccount":           "some-sa",
				"default.repository.serviceaccount.namespace": "some-ns",
			},
		}

		listers := kpacktesthelpers.NewListers([]runtime.Object{kpConfig})
		k8sClient := k8sfakes.NewSimpleClientset(listers.GetKubeObjects()...)
		provider, err := NewKpConfigProvider(context.Background(), k8sClient)
		require.NoError(t, err)
		require.Equal(t, KpConfigProvider{
			client:         k8sClient,
			repoKey:        "default.repository",
			saNameKey:      "default.repository.serviceaccount",
			saNamespaceKey: "default.repository.serviceaccount.namespace",
		}, provider)
	})

	it("uses the old keys in the existing kp config if they exist", func() {
		kpConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"canonical.repository":                          "some-repo",
				"canonical.repository.serviceaccount":           "some-sa",
				"canonical.repository.serviceaccount.namespace": "some-ns",
			},
		}

		listers := kpacktesthelpers.NewListers([]runtime.Object{kpConfig})
		k8sClient := k8sfakes.NewSimpleClientset(listers.GetKubeObjects()...)
		provider, err := NewKpConfigProvider(context.Background(), k8sClient)
		require.NoError(t, err)
		require.Equal(t, KpConfigProvider{
			client:         k8sClient,
			repoKey:        "canonical.repository",
			saNameKey:      "canonical.repository.serviceaccount",
			saNamespaceKey: "canonical.repository.serviceaccount.namespace",
		}, provider)
	})

	it("uses the new keys if the kp config map does not exist", func() {
		client := k8sfakes.NewSimpleClientset()
		provider, err := NewKpConfigProvider(context.Background(), client)
		require.NoError(t, err)
		require.Equal(t, KpConfigProvider{
			client:         client,
			repoKey:        "default.repository",
			saNameKey:      "default.repository.serviceaccount",
			saNamespaceKey: "default.repository.serviceaccount.namespace",
		}, provider)
	})
}
