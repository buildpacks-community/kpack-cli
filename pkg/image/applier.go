package image

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Applier struct {
	KpackClient versioned.Interface
}

func (a *Applier) Apply(image *v1alpha1.Image) error {
	_, err := a.KpackClient.BuildV1alpha1().Images(image.Namespace).Get(image.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	} else if k8serrors.IsNotFound(err) {
		_, err := a.KpackClient.BuildV1alpha1().Images(image.Namespace).Create(image)
		return err
	} else {
		_, err := a.KpackClient.BuildV1alpha1().Images(image.Namespace).Update(image)
		return err
	}
}
