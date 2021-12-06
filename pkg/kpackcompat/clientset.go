package kpackcompat

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

const (
	KpackGroupVersionV1alpha2 = build.GroupName + "/v1alpha2"
)

type ClientsetInterface interface {
	Discovery() discovery.DiscoveryInterface
	KpackV1alpha2() v1alpha2.KpackV1alpha2Interface
}

type ClientSet struct {
	*discovery.DiscoveryClient
	v1alpha2Client v1alpha2.KpackV1alpha2Interface
}

func (c *ClientSet) KpackV1alpha2() v1alpha2.KpackV1alpha2Interface {
	return c.v1alpha2Client
}

func (c *ClientSet) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

func NewForConfig(c *rest.Config) (ClientsetInterface, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs ClientSet
	var err error

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	groups, err := cs.DiscoveryClient.ServerGroups()
	if err != nil {
		return nil, err
	}

	groupVersion, err := GetKpackPreferredGroupVersion(groups)
	if err != nil {
		return nil, err
	}

	if groupVersion == KpackGroupVersionV1alpha2 {
		cs.v1alpha2Client, err = v1alpha2.NewForConfig(&configShallowCopy)
		return &cs, err
	}

	cs.v1alpha2Client, err = NewCompatClient(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func GetKpackPreferredGroupVersion(groups *metav1.APIGroupList) (string, error) {
	for _, g := range groups.Groups {
		if g.Name == build.GroupName {
			return g.PreferredVersion.GroupVersion, nil
		}
	}

	return "", errors.New("kpack.io api group not found")
}
