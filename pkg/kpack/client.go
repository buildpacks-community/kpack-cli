package kpack

import (
	"context"
	"errors"

	"github.com/pivotal/kpack/pkg/apis/build"
	buildV1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	kpackClientSet "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kpackGroup                = build.GroupName
	kpackVersionV1alpha2      = "v1alpha2"
	kpackGroupVersionV1alpha2 = kpackGroup + "/" + kpackVersionV1alpha2
)

var KpackNotFound = errors.New("kpack not found")

type KpackClient interface {
	ListBuilds(ctx context.Context, namespace string, opts v1.ListOptions) (*buildV1alpha2.BuildList, error)
	ListImages(ctx context.Context, namespace string, opts v1.ListOptions) (*buildV1alpha2.ImageList, error)
}

type kpackClient struct {
	client kpackClientSet.Interface

	// groupVersion is the prefered kpack api version
	groupVersion string
}

func NewKpackClient(client kpackClientSet.Interface) (KpackClient, error) {
	groups, err := client.Discovery().ServerGroups()
	if err != nil {
		return nil, err
	}

	groupVersion, err := getKpackGroupVersion(groups)

	return &kpackClient{
		client:       client,
		groupVersion: groupVersion,
	}, nil
}

func getKpackGroupVersion(groups *v1.APIGroupList) (string, error) {
	var foundGroupVersion = ""
	for _, g := range groups.Groups {
		if g.Name == kpackGroup {
			foundGroupVersion = g.PreferredVersion.GroupVersion
		}
	}

	return foundGroupVersion, KpackNotFound
}
