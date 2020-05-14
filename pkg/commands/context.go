package commands

import (
	"os"

	"github.com/pkg/errors"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

type Context struct {
	KpackClient      kpack.Interface
	K8sClient        k8s.Interface
	DefaultNamespace string
}

type ContextProvider interface {
	GetContext() (Context, error)
}

type CommandContextProvider struct {
	context Context
}

func GetContext(contextProvider ContextProvider, namespace *string) (Context, error) {
	context, err := contextProvider.GetContext()
	if err != nil {
		return context, err
	}

	if *namespace == "" {
		*namespace = context.DefaultNamespace
	}
	return context, nil
}

func (c CommandContextProvider) GetContext() (context Context, err error) {
	if c.context.DefaultNamespace, err = getDefaultNamespace(); err != nil {
		return c.context, err
	}

	if c.context.KpackClient, err = getKpackClient(); err != nil {
		return c.context, err
	}

	c.context.K8sClient, err = getK8sClient()
	return c.context, err
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
