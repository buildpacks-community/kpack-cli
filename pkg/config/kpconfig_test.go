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
	ctx := context.Background()
	when("GetKpConfig", func() {
		it("reads from the new keys before the old keys", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository":                            "some-repo",
					"default.repository.serviceaccount":             "some-sa",
					"default.repository.serviceaccount.namespace":   "some-ns",
					"canonical.repository":                          "some-canonical-repo",
					"canonical.repository.serviceaccount":           "some-canonical-sa",
					"canonical.repository.serviceaccount.namespace": "some-canonical-ns",
				},
			}

			listers := kpacktesthelpers.NewListers([]runtime.Object{kpConfig})
			k8sClient := k8sfakes.NewSimpleClientset(listers.GetKubeObjects()...)
			provider := NewKpConfigProvider(k8sClient)
			require.Equal(t, KpConfig{
				defaultRepository: "some-repo",
				serviceAccount:    corev1.ObjectReference{Name: "some-sa", Namespace: "some-ns"},
			}, provider.GetKpConfig(ctx))
		})

		it("reads from the old keys when the new keys don't exist", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"canonical.repository":                          "some-canonical-repo",
					"canonical.repository.serviceaccount":           "some-canonical-sa",
					"canonical.repository.serviceaccount.namespace": "some-canonical-ns",
				},
			}

			listers := kpacktesthelpers.NewListers([]runtime.Object{kpConfig})
			k8sClient := k8sfakes.NewSimpleClientset(listers.GetKubeObjects()...)
			provider := NewKpConfigProvider(k8sClient)
			require.Equal(t, KpConfig{
				defaultRepository: "some-canonical-repo",
				serviceAccount:    corev1.ObjectReference{Name: "some-canonical-sa", Namespace: "some-canonical-ns"},
			}, provider.GetKpConfig(ctx))
		})

		when("SetDefaultRepository", func() {
			it("writes both sets of keys to the config map", func() {
				k8sClient := k8sfakes.NewSimpleClientset()
				provider := NewKpConfigProvider(k8sClient)
				require.NoError(t, provider.SetDefaultRepository(ctx, "some-new-repo"))
				kpConfig, err := k8sClient.CoreV1().ConfigMaps("kpack").Get(ctx, "kp-config", metav1.GetOptions{})
				require.NoError(t, err)
				require.Equal(t, map[string]string{
					"canonical.repository": "some-new-repo",
					"default.repository":   "some-new-repo",
				}, kpConfig.Data)
			})
		})

		when("SetDefaultServiceAccount", func() {
			it("writes both sets of keys to the config map", func() {
				k8sClient := k8sfakes.NewSimpleClientset()
				provider := NewKpConfigProvider(k8sClient)
				require.NoError(t, provider.SetDefaultServiceAccount(ctx, corev1.ObjectReference{
					Name:      "some-new-sa",
					Namespace: "some-new-ns",
				}))
				kpConfig, err := k8sClient.CoreV1().ConfigMaps("kpack").Get(ctx, "kp-config", metav1.GetOptions{})
				require.NoError(t, err)
				require.Equal(t, map[string]string{
					"canonical.repository.serviceaccount":           "some-new-sa",
					"canonical.repository.serviceaccount.namespace": "some-new-ns",
					"default.repository.serviceaccount":             "some-new-sa",
					"default.repository.serviceaccount.namespace":   "some-new-ns",
				}, kpConfig.Data)
			})
		})
	})
}
