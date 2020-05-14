package testhelpers

import (
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

type FakeContextProvider struct {
	context commands.Context
}

func (f FakeContextProvider) GetContext() (commands.Context, error) {
	return f.context, nil
}

func NewFakeContextProvider(
	defaultNamespace string,
	kpackClient *kpackfakes.Clientset,
	k8sClient *k8sfakes.Clientset) FakeContextProvider {

	return FakeContextProvider{
		context: commands.Context{
			KpackClient:      kpackClient,
			K8sClient:        k8sClient,
			DefaultNamespace: defaultNamespace,
		},
	}
}

func NewFakeClusterContextProvider(
	kpackClient *kpackfakes.Clientset,
	k8sClient *k8sfakes.Clientset) FakeContextProvider {

	return FakeContextProvider{
		context: commands.Context{
			KpackClient: kpackClient,
			K8sClient:   k8sClient,
		},
	}
}

func NewFakeKpackContextProvider(
	defaultNamespace string,
	kpackClient *kpackfakes.Clientset) FakeContextProvider {

	return FakeContextProvider{
		context: commands.Context{
			KpackClient:      kpackClient,
			DefaultNamespace: defaultNamespace,
		},
	}
}

func NewFakeKpackClusterContextProvider(
	kpackClient *kpackfakes.Clientset) FakeContextProvider {

	return FakeContextProvider{
		context: commands.Context{
			KpackClient: kpackClient,
		},
	}
}

func NewFakeK8sContextProvider(
	defaultNamespace string,
	k8sClient *k8sfakes.Clientset) FakeContextProvider {

	return FakeContextProvider{
		context: commands.Context{
			K8sClient:        k8sClient,
			DefaultNamespace: defaultNamespace,
		},
	}
}

func NewFakeK8sClusterContextProvider(
	k8sClient *k8sfakes.Clientset) FakeContextProvider {

	return FakeContextProvider{
		context: commands.Context{
			K8sClient: k8sClient,
		},
	}
}
