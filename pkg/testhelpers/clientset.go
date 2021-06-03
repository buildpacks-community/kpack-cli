// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

type FakeClientSetProvider struct {
	clientSet k8s.ClientSet
}

func (f FakeClientSetProvider) GetClientSet(namespace string) (clientSet k8s.ClientSet, err error) {
	if namespace != "" {
		f.clientSet.Namespace = namespace
	}
	return f.clientSet, nil
}

func GetFakeKpackProvider(kpackClient *kpackfakes.Clientset, namespace string) FakeClientSetProvider {
	return FakeClientSetProvider{
		clientSet: k8s.ClientSet{
			KpackClient: kpackClient,
			Namespace:   namespace,
		},
	}
}

func GetFakeKpackClusterProvider(kpackClient *kpackfakes.Clientset) FakeClientSetProvider {
	return FakeClientSetProvider{
		clientSet: k8s.ClientSet{
			KpackClient: kpackClient,
		},
	}
}

func GetFakeK8sProvider(k8sClient *k8sfakes.Clientset, namespace string) FakeClientSetProvider {
	return FakeClientSetProvider{
		clientSet: k8s.ClientSet{
			K8sClient: k8sClient,
			Namespace: namespace,
		},
	}
}

func GetFakeClusterProvider(k8sClient *k8sfakes.Clientset, kpackClient *kpackfakes.Clientset) FakeClientSetProvider {
	return FakeClientSetProvider{
		clientSet: k8s.ClientSet{
			K8sClient:   k8sClient,
			KpackClient: kpackClient,
		},
	}
}
