package commands

import (
	"os"

	"github.com/pkg/errors"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

type ContextProvider interface {
	Initialize() error
	KpackClient() kpack.Interface
	K8sClient() k8s.Interface
	DefaultNamespace() string
}

type CommandContext struct {
	kpackClient      *kpack.Clientset
	k8sClient        *k8s.Clientset
	defaultNamespace string
}

func InitContext(cmdContext ContextProvider, namespace *string) error {
	if err := cmdContext.Initialize(); err != nil {
		return err
	}

	if *namespace == "" {
		*namespace = cmdContext.DefaultNamespace()
	}

	return nil
}

func (c CommandContext) Initialize() (err error) {
	if c.defaultNamespace, err = getDefaultNamespace(); err != nil {
		return err
	}

	if c.kpackClient, err = getKpackClient(); err != nil {
		return err
	}

	c.k8sClient, err = getK8sClient()
	return err
}

func (c CommandContext) KpackClient() kpack.Interface {
	return c.kpackClient
}

func (c CommandContext) K8sClient() k8s.Interface {
	return c.k8sClient
}

func (c CommandContext) DefaultNamespace() string {
	return c.defaultNamespace
}

func getKpackClient() (*kpack.Clientset, error) {
	restConfig, err := restConfig()
	if err != nil {
		return nil, err
	}

	return kpack.NewForConfig(restConfig)
}

func getK8sClient() (*k8s.Clientset, error) {
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

func getDefaultNamespace() (string, error) {
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
