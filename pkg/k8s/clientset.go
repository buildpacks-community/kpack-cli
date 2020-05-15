package k8s

import (
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	k8s "k8s.io/client-go/kubernetes"
)

type ClientSet struct {
	KpackClient kpack.Interface
	K8sClient   k8s.Interface
	Namespace   string
}

type ClientSetProvider struct {
	initializer ClientSetInitializer
}

func NewDefaultClientSetProvider() ClientSetProvider {
	return ClientSetProvider{DefaultClientSetInitializer{}}
}

func NewClientSetProvider(initializer ClientSetInitializer) ClientSetProvider {
	return ClientSetProvider{initializer}
}

func (c ClientSetProvider) GetClientSet(namespace string) (cs ClientSet, err error) {
	if namespace == "" {
		if cs.Namespace, err = c.initializer.GetDefaultNamespace(); err != nil {
			return
		}
	} else {
		cs.Namespace = namespace
	}

	if cs.KpackClient, err = c.initializer.GetKpackClient(); err != nil {
		return
	}

	cs.K8sClient, err = c.initializer.GetK8sClient()
	return
}
