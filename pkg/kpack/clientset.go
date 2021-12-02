package kpack

import (
	"fmt"
	"os"

	kpackv1alpha1 "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackv1alpha2 "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/util/flowcontrol"
)

type ClientSet struct {
	KpackClient kpack.Interface
	Namespace   string
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

	return d.clientSet, err
}

func (d DefaultClientSetProvider) getKpackClient() (*kpack.Clientset, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return kpack.NewForConfig(restConfig)
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

type ClientSetWrapper struct {
	v1alpha2Client kpackv1alpha2.KpackV1alpha2Interface
}

// KpackV1alpha1 retrieves the KpackV1alpha1Client
func (c *ClientSetWrapper) KpackV1alpha1() kpackv1alpha1.KpackV1alpha1Interface {
	return nil
}

// KpackV1alpha2 retrieves the KpackV1alpha2Client
func (c *ClientSetWrapper) KpackV1alpha2() kpackv1alpha2.KpackV1alpha2Interface {
	return c.v1alpha2Client
}

// Discovery retrieves the DiscoveryClient
func (c *ClientSetWrapper) Discovery() discovery.DiscoveryInterface {
	return nil
}

func NewForConfig(c *rest.Config) (kpack.Interface, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs ClientSetWrapper
	var err error

	if c.GroupVersion.String() == kpackGroupVersionV1alpha2 {
		cs.v1alpha2Client, err = kpackv1alpha2.NewForConfig(&configShallowCopy)
		return &cs, err
	}

	cs.v1alpha2Client, err = NewBuildClientForConfig(&configShallowCopy)
	return &cs, err
}
