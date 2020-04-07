package image

import (
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deleter struct {
	KpackClient versioned.Interface
}

func (d *Deleter) Delete(namespace, name string) error {
	return d.KpackClient.BuildV1alpha1().Images(namespace).Delete(name, &metav1.DeleteOptions{})
}
