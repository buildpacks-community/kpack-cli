package image

import (
	"fmt"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"strings"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

type filter struct {
	filterFunc func(image v1alpha1.Image, values []string) bool
	values     []string
}

func filterImageList(images *v1alpha1.ImageList, flags []string) (*v1alpha1.ImageList, error) {
	filters, err := parseFilters(flags)
	if err != nil {
		return nil, err
	}

	if len(filters) == 0 {
		return images, nil
	}

	var filteredItems []v1alpha1.Image
	for _, item := range images.Items {
		if matchesAll(item, filters) {
			filteredItems = append(filteredItems, item)
		}
	}

	images.Items = filteredItems
	return images, nil
}

func parseFilters(flags []string) ([]filter, error) {
	var (
		filters             []filter
		builderRegex        = regexp.MustCompile(`^builder=(.+)$`)
		clusterBuilderRegex = regexp.MustCompile(`^clusterbuilder=(.+)$`)
		latestReasonRegex   = regexp.MustCompile(`^latest-reason=(.+)$`)
		statusRegex         = regexp.MustCompile(`^ready=(.+)$`)
	)

	for _, flag := range flags {
		m := builderRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{values: m[1:], filterFunc: func(image v1alpha1.Image, values []string) bool {
				return image.Spec.Builder.Kind == v1alpha1.BuilderKind && image.Spec.Builder.Name == values[0]
			}})
			continue
		}

		m = clusterBuilderRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{values: m[1:], filterFunc: func(image v1alpha1.Image, values []string) bool {
				return image.Spec.Builder.Kind == v1alpha1.ClusterBuilderKind && image.Spec.Builder.Name == values[0]
			}})
			continue
		}

		m = latestReasonRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{values: strings.Split(m[1], ","), filterFunc: matchesLatestReason})
			continue
		}

		m = statusRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{values: strings.Split(m[1], ","), filterFunc: matchesStatus})
			continue
		}

		return nil, fmt.Errorf(`invalid filter argument "%s"`, flag)
	}

	return filters, nil
}

func matchesAll(image v1alpha1.Image, fs []filter) bool {
	for _, f := range fs {
		if !f.filterFunc(image, f.values) {
			return false
		}
	}

	return true
}

func matchesStatus(image v1alpha1.Image, values []string) bool {
	contains := func(q corev1.ConditionStatus) bool {
		for _, v := range values {
			if strings.ToLower(string(q)) == strings.ToLower(v) {
				return true
			}
		}

		return false
	}

	cond := image.Status.GetCondition(corev1alpha1.ConditionReady)

	switch cond {
	case nil:
		return contains(corev1.ConditionUnknown)
	default:
		return contains(cond.Status)
	}
}

func matchesLatestReason(image v1alpha1.Image, values []string) bool {
	for _, v := range values {
		for _, reason := range strings.Split(image.Status.LatestBuildReason, ",") {
			if strings.ToLower(reason) == strings.ToLower(v) {
				return true
			}
		}
	}

	return false
}
