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

func TestDefaultRepositoryCommand(t *testing.T) {
	spec.Run(t, "TestDefaultRepositoryCommand", testDefaultRepositoryCommand)
}

func testDefaultRepositoryCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, _ *kpackfakes.Clientset) *cobra.Command {
		return NewDefaultRepositoryCommand(testhelpers.GetFakeClusterProvider(k8sClientSet, nil))
	}

	when("running command without any args", func() {
		it("prints the current default repository value when it is not empty", func() {
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
				Args:                []string{},
				ExpectedOutput:      "test-repo\n",
				ExpectedErrorOutput: "",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("prints an error when default-repository field is empty", func() {
			testhelpers.CommandTest{
				Objects:             []runtime.Object{},
				Args:                []string{},
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("setting the default repository", func() {
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
				Args:                []string{"new-repo"},
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
								"canonical.repository":              "new-repo",
								"default.repository":                "new-repo",
								"default.repository.serviceaccount": "default",
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
				ExpectedOutput:      "kp-config set\n",
				ExpectedErrorOutput: "",
				ExpectCreates: []runtime.Object{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kp-config",
							Namespace: "kpack",
						},
						Data: map[string]string{
							"default.repository":   "new-repo",
							"canonical.repository": "new-repo",
						},
					},
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})
}
