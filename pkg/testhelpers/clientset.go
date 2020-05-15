package testhelpers

import (
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

type FakeClientSetInitializer struct {
	KpackClient kpack.Interface
	K8sClient   kubernetes.Interface
	Namespace   string
}

func (f FakeClientSetInitializer) GetKpackClient() (kpack.Interface, error) {
	return f.KpackClient, nil
}

func (f FakeClientSetInitializer) GetK8sClient() (kubernetes.Interface, error) {
	return f.K8sClient, nil
}

func (f FakeClientSetInitializer) GetDefaultNamespace() (string, error) {
	return f.Namespace, nil
}

func GetFakeKpackProvider(
	kpackClient *kpackfakes.Clientset,
	namespace string) k8s.ClientSetProvider {

	return k8s.NewClientSetProvider(
		FakeClientSetInitializer{
			KpackClient: kpackClient,
			Namespace:   namespace,
		},
	)
}

func GetFakeKpackClusterProvider(
	kpackClient *kpackfakes.Clientset) k8s.ClientSetProvider {

	return k8s.NewClientSetProvider(
		FakeClientSetInitializer{
			KpackClient: kpackClient,
		},
	)
}

func GetFakeK8sProvider(
	k8sClient *k8sfakes.Clientset,
	namespace string) k8s.ClientSetProvider {

	return k8s.NewClientSetProvider(
		FakeClientSetInitializer{
			K8sClient: k8sClient,
			Namespace: namespace,
		},
	)
}
