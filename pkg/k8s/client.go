package k8s

import (
	"os"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKpackClient() (versioned.Interface, error) {
	clusterConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	restConfig, err := clusterConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return versioned.NewForConfig(restConfig)
}
