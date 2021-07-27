package config

import (
	"testing"

	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestCanonicalRepositoryCommand(t *testing.T) {
	spec.Run(t, "TestCanonicalRepositoryCommand", testCanonicalRepositoryCommand)
}

func testCanonicalRepositoryCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, _ *kpackfakes.Clientset) *cobra.Command {
		return NewCanonicalRepositoryCommand(testhelpers.GetFakeClusterProvider(k8sClientSet, nil))
	}

	when("running command without any args", func() {
		it("prints the current canonical repository value when it is not empty", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"canonical.repository":                "test-repo",
					"canonical.repository.serviceaccount": "default",
				},
			}

			testhelpers.CommandTest{
				Objects:             []runtime.Object{kpConfig},
				Args:                []string{},
				ExpectErr:           false,
				ExpectedOutput:      "test-repo\n",
				ExpectedErrorOutput: "",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("prints an error when canonical-repository field is empty", func() {
			testhelpers.CommandTest{
				Objects:        []runtime.Object{},
				Args:           []string{},
				ExpectErr:      true,
				ExpectedOutput: "Error: failed to get canonical repository: use \"kp config canonical-repository\" to set\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("setting the canonical repository", func() {
		it("updates the existing config map if it exists", func() {
			kpConfig := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kp-config",
					Namespace: "kpack",
				},
				Data: map[string]string{
					"canonical.repository":                "test-repo",
					"canonical.repository.serviceaccount": "default",
				},
			}

			testhelpers.CommandTest{
				Objects:             []runtime.Object{kpConfig},
				Args:                []string{"new-repo"},
				ExpectErr:           false,
				ExpectedOutput:      "kp-config set\n",
				ExpectedErrorOutput: "",
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "kp-config",
								Namespace: "kpack",
							},
							Data: map[string]string{
								"canonical.repository":                "new-repo",
								"canonical.repository.serviceaccount": "default",
							},
						},
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("creates a new config map if it doesn't exist ", func() {
			testhelpers.CommandTest{
				Objects:             []runtime.Object{},
				Args:                []string{"new-repo"},
				ExpectErr:           false,
				ExpectedOutput:      "kp-config set\n",
				ExpectedErrorOutput: "",
				ExpectCreates: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kp-config",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"canonical.repository":                          "new-repo",
							"canonical.repository.serviceaccount":           "",
							"canonical.repository.serviceaccount.namespace": "",
						},
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})
}
