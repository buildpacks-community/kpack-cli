package k8s

import (
	"os"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKpackClient() (kpack.Interface, error) {
	restConfig, err := restConfig()
	if err != nil {
		return nil, err
	}

	return kpack.NewForConfig(restConfig)
}

func NewK8sClient() (k8s.Interface, error) {
	restConfig, err := restConfig()
	if err != nil {
		return nil, err
	}

	return k8s.NewForConfig(restConfig)
}

func restConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	restConfig, err := clientConfig.ClientConfig()
	return restConfig, err
}
