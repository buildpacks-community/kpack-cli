package kpackcompat

import (
	kpackv1alpha1 "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	kpackV1alpha2 "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	"k8s.io/client-go/rest"
)

type KpackV1alpha1CompatClient struct {
	v1alpha1KpackClient kpackv1alpha1.KpackV1alpha1Interface
}

func (c *KpackV1alpha1CompatClient) Builds(namespace string) kpackV1alpha2.BuildInterface {
	return newBuilds(c, namespace)
}

func (c *KpackV1alpha1CompatClient) Builders(namespace string) kpackV1alpha2.BuilderInterface {
	return newBuilders(c, namespace)
}

func (c *KpackV1alpha1CompatClient) ClusterBuilders() kpackV1alpha2.ClusterBuilderInterface {
	return newClusterBuilders(c)
}

func (c *KpackV1alpha1CompatClient) ClusterStacks() kpackV1alpha2.ClusterStackInterface {
	return newClusterStacks(c)
}

func (c *KpackV1alpha1CompatClient) ClusterStores() kpackV1alpha2.ClusterStoreInterface {
	return newClusterStores(c)
}

func (c *KpackV1alpha1CompatClient) Images(namespace string) kpackV1alpha2.ImageInterface {
	return newImages(c, namespace)
}

func (c *KpackV1alpha1CompatClient) SourceResolvers(namespace string) kpackV1alpha2.SourceResolverInterface {
	return newSourceResolvers(c, namespace)
}

func NewCompatClient(c *rest.Config) (*KpackV1alpha1CompatClient, error) {
	client, err := kpackv1alpha1.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &KpackV1alpha1CompatClient{client}, nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *KpackV1alpha1CompatClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.v1alpha1KpackClient.RESTClient()
}
