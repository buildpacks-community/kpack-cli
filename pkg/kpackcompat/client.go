package kpackcompat

import (
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	"k8s.io/client-go/rest"
)

const LatestKpackAPIVersion = "v1alpha2"

type kpackV1alpha1CompatClient struct {
	v1alpha1KpackClient v1alpha1.KpackV1alpha1Interface
}

func (c *kpackV1alpha1CompatClient) Builds(namespace string) v1alpha2.BuildInterface {
	return newBuilds(c, namespace)
}

func (c *kpackV1alpha1CompatClient) Builders(namespace string) v1alpha2.BuilderInterface {
	return newBuilders(c, namespace)
}

func (c *kpackV1alpha1CompatClient) ClusterBuilders() v1alpha2.ClusterBuilderInterface {
	return newClusterBuilders(c)
}

func (c *kpackV1alpha1CompatClient) ClusterStacks() v1alpha2.ClusterStackInterface {
	return newClusterStacks(c)
}

func (c *kpackV1alpha1CompatClient) ClusterStores() v1alpha2.ClusterStoreInterface {
	return newClusterStores(c)
}

func (c *kpackV1alpha1CompatClient) Images(namespace string) v1alpha2.ImageInterface {
	return newImages(c, namespace)
}

func (c *kpackV1alpha1CompatClient) SourceResolvers(namespace string) v1alpha2.SourceResolverInterface {
	return newSourceResolvers(c, namespace)
}

func newV1Alpha1CompatClient(c versioned.Interface) *kpackV1alpha1CompatClient {
	return &kpackV1alpha1CompatClient{c.KpackV1alpha1()}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *kpackV1alpha1CompatClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.v1alpha1KpackClient.RESTClient()
}
