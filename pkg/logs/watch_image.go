package logs

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/kpack-cli/pkg/kpackcompat"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
)

type watchOneImage struct {
	kpackClient kpackcompat.ClientsetInterface
	image       *v1alpha2.Image
	ctx         context.Context
}

func (w watchOneImage) Watch(options v1.ListOptions) (watch.Interface, error) {
	options.FieldSelector = fmt.Sprintf("metadata.name=%s", w.image.Name)
	return w.kpackClient.KpackV1alpha2().Images(w.image.Namespace).Watch(w.ctx, options)
}
