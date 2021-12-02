package kpack

import (
	"github.com/pivotal/kpack/pkg/apis/build"
	v1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	kpackV1alpha2 "github.com/pivotal/kpack/pkg/client/clientset/versioned/typed/build/v1alpha2"
	rest "k8s.io/client-go/rest"
)

const (
	kpackGroup                = build.GroupName
	kpackVersionV1alpha2      = "v1alpha2"
	kpackGroupVersionV1alpha2 = kpackGroup + "/" + kpackVersionV1alpha2
)

// KpackV1alpha2Client is used to interact with features provided by the kpack.io group.
type KpackV1alpha2Client struct {
	restClient rest.Interface
}

func (c *KpackV1alpha2Client) Builds(namespace string) kpackV1alpha2.BuildInterface {
	return newBuilds(c, namespace)
}

func (c *KpackV1alpha2Client) Builders(namespace string) kpackV1alpha2.BuilderInterface {
	return newBuilders(c, namespace)
}

func (c *KpackV1alpha2Client) ClusterBuilders() kpackV1alpha2.ClusterBuilderInterface {
	return newClusterBuilders(c)
}

func (c *KpackV1alpha2Client) ClusterStacks() kpackV1alpha2.ClusterStackInterface {
	return newClusterStacks(c)
}

func (c *KpackV1alpha2Client) ClusterStores() kpackV1alpha2.ClusterStoreInterface {
	return newClusterStores(c)
}

func (c *KpackV1alpha2Client) Images(namespace string) kpackV1alpha2.ImageInterface {
	return newImages(c, namespace)
}

func (c *KpackV1alpha2Client) SourceResolvers(namespace string) kpackV1alpha2.SourceResolverInterface {
	return newSourceResolvers(c, namespace)
}

// NewForConfig creates a new KpackV1alpha2Client for the given config.
func NewBuildClientForConfig(c *rest.Config) (*KpackV1alpha2Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &KpackV1alpha2Client{client}, nil
}

// New creates a new KpackV1alpha2Client for the given RESTClient.
func New(c rest.Interface) *KpackV1alpha2Client {
	return &KpackV1alpha2Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha2.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *KpackV1alpha2Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
