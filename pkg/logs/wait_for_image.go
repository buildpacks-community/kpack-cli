package logs

import (
	"context"
	"fmt"
	"io"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/kpack-cli/pkg/kpackcompat"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchTools "k8s.io/client-go/tools/watch"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

type imageWaiter struct {
	KpackClient kpackcompat.ClientsetInterface
	logTailer   ImageLogTailer
}

type ImageLogTailer interface {
	TailBuildName(ctx context.Context, writer io.Writer, buildName, namespace string) error
}

func NewImageWaiter(kpackClient kpackcompat.ClientsetInterface, logTailer ImageLogTailer) *imageWaiter {
	return &imageWaiter{KpackClient: kpackClient, logTailer: logTailer}
}

func (w *imageWaiter) Wait(ctx context.Context, writer io.Writer, image *v1alpha2.Image) (string, error) {
	if done, err := imageUpdateHasResolved(ctx, image.Generation)(watch.Event{Object: image}); err != nil {
		return "", err
	} else if done {
		return w.resultOfImageWait(ctx, writer, image.Generation, image)
	}

	event, err := watchTools.Until(ctx,
		image.ResourceVersion,
		watchOneImage{kpackClient: w.KpackClient, image: image, ctx: ctx},
		filterErrors(imageUpdateHasResolved(ctx, image.Generation)))
	if err != nil {
		return "", err
	}

	i, err := eventToV1alpha2Image(ctx, event)
	if err != nil {
		return "", err
	}

	return w.resultOfImageWait(ctx, writer, image.Generation, i)
}

func eventToV1alpha2Image(ctx context.Context, event *watch.Event) (*v1alpha2.Image, error) {
	image := &v1alpha2.Image{}
	switch event.Object.(type) {
	case *v1alpha2.Image:
		image = event.Object.(*v1alpha2.Image)
	case *v1alpha1.Image:
		v1Image := event.Object.(*v1alpha1.Image)
		err := image.ConvertFrom(ctx, v1Image)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unexpected object received")
	}
	return image, nil
}

func imageUpdateHasResolved(ctx context.Context, generation int64) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		image, err := eventToV1alpha2Image(ctx, &event)
		if err != nil {
			return false, err
		}

		if image.Status.ObservedGeneration == generation { // image is reconciled
			if !image.Status.GetCondition(corev1alpha1.ConditionReady).IsUnknown() {
				return true, nil // image is resolved
			} else if image.Status.LatestBuildImageGeneration == generation {
				return true, nil // Build scheduled
			} else {
				return false, nil // still waiting on build to be scheduled
			}
		} else if image.Status.ObservedGeneration > generation {
			return false, errors.Errorf("image %s was updated before original update was processed", image.Name) // update skipped
		} else {
			return false, nil // still waiting on update
		}
	}
}

func (w *imageWaiter) resultOfImageWait(ctx context.Context, writer io.Writer, generation int64, image *v1alpha2.Image) (string, error) {
	if image.Status.LatestBuildImageGeneration == generation {
		return w.waitBuild(ctx, writer, image.Namespace, image.Status.LatestBuildRef)
	}

	if condition := image.Status.GetCondition(corev1alpha1.ConditionReady); condition.IsFalse() {
		return "", imageFailure(image.Name, condition.Message)
	}

	return image.Status.LatestImage, nil
}

func imageFailure(name, statusMessage string) error {
	errMsg := fmt.Sprintf("update to image %s failed", name)

	if statusMessage != "" {
		errMsg = fmt.Sprintf("%s: %s", errMsg, statusMessage)
	}
	return errors.New(errMsg)
}

func (w *imageWaiter) waitBuild(ctx context.Context, writer io.Writer, namespace, buildName string) (string, error) {
	doneChan := make(chan struct{})
	defer func() { <-doneChan }()

	go func() { // tail logs
		defer close(doneChan)
		err := w.logTailer.TailBuildName(ctx, writer, namespace, buildName)
		if err != nil {
			fmt.Fprintf(writer, "error tailing logs %s", err)
		}
	}()

	build, err := w.buildWatchUntil(ctx, namespace, buildName, filterErrors(buildHasResolved(ctx)))
	if err != nil {
		return "", err
	}

	if condition := build.Status.GetCondition(corev1alpha1.ConditionSucceeded); condition.IsFalse() {
		return "", buildFailure(condition.Message)
	}

	return build.Status.LatestImage, nil
}

func buildHasResolved(ctx context.Context) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		build, err := eventToV1alpha2Build(ctx, &event)
		if err != nil {
			return false, errors.New("unexpected object received, expected Build")
		}

		return !build.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsUnknown(), nil
	}
}

func eventToV1alpha2Build(ctx context.Context, event *watch.Event) (*v1alpha2.Build, error) {
	build := &v1alpha2.Build{}
	switch event.Object.(type) {
	case *v1alpha2.Build:
		build = event.Object.(*v1alpha2.Build)
	case *v1alpha1.Build:
		v1Build := event.Object.(*v1alpha1.Build)
		err := build.ConvertFrom(ctx, v1Build)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unexpected object received")
	}
	return build, nil
}

func buildFailure(statusMessage string) error {
	errMsg := "build failed"

	if statusMessage != "" {
		errMsg = fmt.Sprintf("%s: %s", errMsg, statusMessage)
	}
	return errors.New(errMsg)
}

func (w *imageWaiter) buildWatchUntil(ctx context.Context, namespace, buildName string, condition watchTools.ConditionFunc) (*v1alpha2.Build, error) {
	build, err := w.KpackClient.KpackV1alpha2().Builds(namespace).Get(ctx, buildName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var buildType runtime.Object = &v1alpha1.Build{}
	groups, err := w.KpackClient.Discovery().ServerGroups()
	if err != nil {
		return nil, err
	}

	groupVersion, err := kpackcompat.GetKpackPreferredGroupVersion(groups)
	if err != nil {
		return nil, err
	}

	if groupVersion == kpackcompat.KpackGroupVersionV1alpha2 {
		buildType = &v1alpha2.Build{}
	}

	event, err := watchTools.UntilWithSync(ctx,
		&watchOneBuild{context: ctx, kpackClient: w.KpackClient, namespace: namespace, buildName: buildName},
		buildType,
		func(store cache.Store) (bool, error) {
			return condition(watch.Event{Object: build})
		},
		condition,
	)
	if err != nil {
		return nil, err
	}
	if event != nil { // event is nil if precondition is true
		build, err = eventToV1alpha2Build(ctx, event)
		if err != nil {
			return nil, errors.New("unexpected object received, expected Build")
		}
	}
	return build, nil
}

func filterErrors(condition watchTools.ConditionFunc) watchTools.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Error {
			return false, errors.Errorf("error on watch %+v", event.Object)
		}

		return condition(event)
	}
}
