// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"os"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/buildpacks-community/kpack-cli/pkg/kpackcompat"
)

type ClientSet struct {
	KpackClient   versioned.Interface
	K8sClient     k8s.Interface
	DynamicClient dynamic.Interface
	Namespace     string
}

type ClientSetProvider interface {
	GetClientSet(namespace string) (ClientSet, error)
}

type DefaultClientSetProvider struct {
	clientSet ClientSet
}

func (d DefaultClientSetProvider) GetClientSet(namespace string) (ClientSet, error) {
	var err error

	if namespace == "" {
		if d.clientSet.Namespace, err = d.getDefaultNamespace(); err != nil {
			return d.clientSet, err
		}
	} else {
		d.clientSet.Namespace = namespace
	}

	if d.clientSet.KpackClient, err = d.getKpackClient(); err != nil {
		return d.clientSet, err
	}

	if d.clientSet.DynamicClient, err = d.getDynamicClient(); err != nil {
		return d.clientSet, err
	}

	d.clientSet.K8sClient, err = d.getK8sClient()
	return d.clientSet, err
}

func (d DefaultClientSetProvider) getKpackClient() (versioned.Interface, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	client, err := kpackcompat.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (d DefaultClientSetProvider) getK8sClient() (*k8s.Clientset, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return k8s.NewForConfig(restConfig)
}

func (d DefaultClientSetProvider) getDynamicClient() (dynamic.Interface, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(restConfig)
}

func (d DefaultClientSetProvider) restConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	restConfig, err := clientConfig.ClientConfig()
	return restConfig, err
}

func (d DefaultClientSetProvider) getDefaultNamespace() (string, error) {
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
