package testhelpers

import (
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	k8s "k8s.io/client-go/kubernetes"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
)

type FakeContext struct {
	defaultNamespace string
	kpackClient      *kpackfakes.Clientset
	k8sClient        *k8sfakes.Clientset
}

func (f FakeContext) Initialize() error            { return nil }
func (f FakeContext) KpackClient() kpack.Interface { return f.kpackClient }
func (f FakeContext) K8sClient() k8s.Interface     { return f.k8sClient }
func (f FakeContext) DefaultNamespace() string     { return f.defaultNamespace }

func NewFakeContext(
	defaultNamespace string,
	kpackClient *kpackfakes.Clientset,
	k8sClient *k8sfakes.Clientset) FakeContext {
	return FakeContext{
		defaultNamespace: defaultNamespace,
		kpackClient:      kpackClient,
		k8sClient:        k8sClient,
	}
}

func NewFakeClusterContext(
	kpackClient *kpackfakes.Clientset,
	k8sClient *k8sfakes.Clientset) FakeContext {
	return FakeContext{
		kpackClient: kpackClient,
		k8sClient:   k8sClient,
	}
}

func NewFakeKpackContext(
	defaultNamespace string,
	kpackClient *kpackfakes.Clientset) FakeContext {
	return FakeContext{
		defaultNamespace: defaultNamespace,
		kpackClient:      kpackClient,
	}
}

func NewFakeKpackClusterContext(
	kpackClient *kpackfakes.Clientset) FakeContext {
	return FakeContext{
		kpackClient: kpackClient,
	}
}

func NewFakeK8sContext(
	defaultNamespace string,
	k8sClient *k8sfakes.Clientset) FakeContext {
	return FakeContext{
		defaultNamespace: defaultNamespace,
		k8sClient:        k8sClient,
	}
}

func NewFakeK8sClusterContext(
	k8sClient *k8sfakes.Clientset) FakeContext {
	return FakeContext{
		k8sClient: k8sClient,
	}
}
