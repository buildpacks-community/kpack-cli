package logs

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/kpack-cli/pkg/kpackcompat"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type watchOneBuild struct {
	buildName   string
	kpackClient kpackcompat.ClientsetInterface
	namespace   string
	context     context.Context
}

func (l *watchOneBuild) Watch(options v1.ListOptions) (watch.Interface, error) {
	options.FieldSelector = fmt.Sprintf("metadata.name=%s", l.buildName)

	return l.kpackClient.KpackV1alpha2().Builds(l.namespace).Watch(l.context, options)
}

func (l *watchOneBuild) List(options v1.ListOptions) (runtime.Object, error) {
	options.FieldSelector = fmt.Sprintf("metadata.name=%s", l.buildName)

	return l.kpackClient.KpackV1alpha2().Builds(l.namespace).List(l.context, options)
}
