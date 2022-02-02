package kpackcompat

import (
	"github.com/pivotal/kpack/pkg/apis/build"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

const (
	kpackGroupVersionV1alpha2 = build.GroupName + "/v1alpha2"
)

type ClientSetWrapper struct {
	kpackClientSet versioned.Interface
}

func (c *ClientSetWrapper) Discovery() discovery.DiscoveryInterface {
	return c.kpackClientSet.Discovery()
}

func (c *ClientSetWrapper) KpackV1alpha1() v1alpha1.KpackV1alpha1Interface {
	return c.kpackClientSet.KpackV1alpha1()
}

func (c *ClientSetWrapper) KpackV1alpha2() v1alpha2.KpackV1alpha2Interface {
	return newV1Alpha1CompatClient(c.kpackClientSet)
}


func NewForConfig(c *rest.Config) (versioned.Interface, error) {
	realClientSet, err := versioned.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	groups, err := realClientSet.Discovery().ServerGroups()
	if err != nil {
		return nil, err
	}

	groupVersion, err := getKpackPreferredGroupVersion(groups)
	if err != nil {
		return nil, err
	}

	if groupVersion == kpackGroupVersionV1alpha2 {
		return realClientSet, nil
	}

	return &ClientSetWrapper{kpackClientSet: realClientSet}, nil
}

func getKpackPreferredGroupVersion(groups *metav1.APIGroupList) (string, error) {
	for _, g := range groups.Groups {
		if g.Name == build.GroupName {
			return g.PreferredVersion.GroupVersion, nil
		}
	}

	return "", errors.New("kpack.io api group not found")
}
