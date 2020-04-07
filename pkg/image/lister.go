package image

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Lister struct {
	KpackClient versioned.Interface
}

func (l *Lister) List(namespace string) (*v1alpha1.ImageList, error) {
	return l.KpackClient.BuildV1alpha1().Images(namespace).List(metav1.ListOptions{})
}
