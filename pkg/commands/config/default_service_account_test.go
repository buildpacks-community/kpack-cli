package config

import (
	"testing"

	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
)

func TestDefaultServiceAccountCommand(t *testing.T) {
	spec.Run(t, "TestDefaultServiceAccountCommand", testDefaultServiceAccountCommand)
}

func testDefaultServiceAccountCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, _ *kpackfakes.Clientset) *cobra.Command {
		return NewDefaultServiceAccountCommand(testhelpers.GetFakeClusterProvider(k8sClientSet, nil))
	}

	when("running command without any args", func() {
		it("prints the current default service account values when it is not empty", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository":                            "test-repo",
					"default.repository.serviceaccount":             "default",
					"default.repository.serviceaccount.namespace":   "default",
					"canonical.repository.serviceaccount":           "default",
					"canonical.repository.serviceaccount.namespace": "default",
				},
			}

			testhelpers.CommandTest{
				Objects:             []runtime.Object{kpConfig},
				Args:                []string{},
				ExpectedOutput:      "Name: default\nNamespace: default\n",
				ExpectedErrorOutput: "",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("it defaults to kpack namespace when default.repository.serviceaccount.namespace is not present", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository":                  "test-repo",
					"default.repository.serviceaccount":   "default",
					"canonical.repository":                "test-repo",
					"canonical.repository.serviceaccount": "default",
				},
			}

			testhelpers.CommandTest{
				Objects:             []runtime.Object{kpConfig},
				Args:                []string{},
				ExpectedOutput:      "Name: default\nNamespace: kpack\n",
				ExpectedErrorOutput: "",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("returns default service account in the kpack namespace when it is empty", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{},
			}

			testhelpers.CommandTest{
				Objects:        []runtime.Object{kpConfig},
				Args:           []string{},
				ExpectedOutput: "Name: default\nNamespace: kpack\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("returns default service account in the kpack namespace when kp-config doesn't exist", func() {
			testhelpers.CommandTest{
				Args:           []string{},
				ExpectedOutput: "Name: default\nNamespace: kpack\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("setting the default service account", func() {
		it("updates the existing config map if it exists", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository":                "test-repo",
					"default.repository.serviceaccount": "default",
				},
			}

			testhelpers.CommandTest{
				Objects:             []runtime.Object{kpConfig},
				Args:                []string{"some-service-account"},
				ExpectedOutput:      "kp-config set\n",
				ExpectedErrorOutput: "",
				ExpectPatches: []string{
					`{"data":{"canonical.repository.serviceaccount":"some-service-account","canonical.repository.serviceaccount.namespace":"kpack","default.repository.serviceaccount":"some-service-account","default.repository.serviceaccount.namespace":"kpack"}}`,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("allows you to set the namespace of the default service account", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"default.repository":                "test-repo",
					"default.repository.serviceaccount": "default",
				},
			}

			testhelpers.CommandTest{
				Objects: []runtime.Object{kpConfig},
				Args: []string{
					"some-service-account",
					"--service-account-namespace",
					"default",
				},
				ExpectedOutput:      "kp-config set\n",
				ExpectedErrorOutput: "",
				ExpectPatches: []string{
					`{"data":{"canonical.repository.serviceaccount":"some-service-account","canonical.repository.serviceaccount.namespace":"default","default.repository.serviceaccount":"some-service-account","default.repository.serviceaccount.namespace":"default"}}`,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("creates a new config map if it doesn't exist ", func() {
			testhelpers.CommandTest{
				Objects:             []runtime.Object{},
				Args:                []string{"some-account"},
				ExpectedOutput:      "kp-config set\n",
				ExpectedErrorOutput: "",
				ExpectCreates: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kp-config",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"default.repository.serviceaccount":             "some-account",
							"default.repository.serviceaccount.namespace":   "kpack",
							"canonical.repository.serviceaccount":           "some-account",
							"canonical.repository.serviceaccount.namespace": "kpack",
						},
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})
}
