package k8s

import (
	"os"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientSetInitializer interface {
	GetKpackClient() (kpack.Interface, error)
	GetK8sClient() (k8s.Interface, error)
	GetDefaultNamespace() (string, error)
}

type DefaultClientSetInitializer struct{}

func (d DefaultClientSetInitializer) GetKpackClient() (kpack.Interface, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return kpack.NewForConfig(restConfig)
}

func (d DefaultClientSetInitializer) GetK8sClient() (k8s.Interface, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return k8s.NewForConfig(restConfig)
}

func (d DefaultClientSetInitializer) restConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	restConfig, err := clientConfig.ClientConfig()
	return restConfig, err
}

func (d DefaultClientSetInitializer) GetDefaultNamespace() (string, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return "", err
	}

	if _, ok := rawConfig.Contexts[rawConfig.CurrentContext]; !ok {
		return "", errors.New("Kubernetes current context is not set")
	}

	defaultNamespace := rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	if defaultNamespace == "" {
		defaultNamespace = "default"
	}

	return defaultNamespace, nil
}
