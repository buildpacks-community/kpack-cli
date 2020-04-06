package k8s

import (
	"errors"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

func GetDefaultNamespace() (string, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return "", err
	}

	_, ok := rawConfig.Contexts[rawConfig.CurrentContext]
	if !ok {
		return "", errors.New("could not get current context")
	}

	namespace := rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	if namespace == "" {
		namespace = "default"
	}

	return namespace, nil
}
