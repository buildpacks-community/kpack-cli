package image

import (
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"strings"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

const (
	filterKindBuilder        = "builder"
	filterKindClusterBuilder = "clusterbuilder"
	filterKindLatestReason   = "latest-reason"
	filterKindStatus         = "ready"
)

type filter struct {
	kind   string
	values []string
}

func Filter(images *v1alpha1.ImageList, flags []string) *v1alpha1.ImageList {
	filters := parseFilters(flags)
	if len(filters) == 0 {
		return images
	}

	var filteredItems []v1alpha1.Image

	for _, item := range images.Items {
		if matchesAll(item, filters) {
			filteredItems = append(filteredItems, item)
		}
	}

	images.Items = filteredItems
	return images
}

func parseFilters(flags []string) []filter {
	var (
		filters             []filter
		builderRegex        = regexp.MustCompile(`^builder=(.*)$`)
		clusterBuilderRegex = regexp.MustCompile(`^clusterbuilder=(.*)$`)
		latestReasonRegex   = regexp.MustCompile(`^latest-reason=(.*)$`)
		statusRegex         = regexp.MustCompile(`^ready=(.*)$`)
	)

	for _, flag := range flags {
		m := builderRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{kind: filterKindBuilder, values: m[1:]})
		}

		m = clusterBuilderRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{kind: filterKindClusterBuilder, values: m[1:]})
		}

		m = latestReasonRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{kind: filterKindLatestReason, values: strings.Split(m[1], ",")})
		}

		m = statusRegex.FindStringSubmatch(flag)
		if len(m) == 2 {
			filters = append(filters, filter{kind: filterKindStatus, values: strings.Split(m[1], ",")})
		}
	}

	return filters
}

func matchesAll(image v1alpha1.Image, fs []filter) bool {
	for _, f := range fs {
		if !matches(image, f) {
			return false
		}
	}

	return true
}

func matches(image v1alpha1.Image, f filter) bool {
	switch f.kind {
	case filterKindBuilder:
		if image.Spec.Builder.Kind == v1alpha1.BuilderKind && image.Spec.Builder.Name == f.values[0] {
			return true
		}
	case filterKindClusterBuilder:
		if image.Spec.Builder.Kind == v1alpha1.ClusterBuilderKind && image.Spec.Builder.Name == f.values[0] {
			return true
		}
	case filterKindStatus:
		if matchesStatus(image, f.values) {
			return true
		}
	case filterKindLatestReason:
		if matchesLatestReason(image, f.values) {
			return true
		}
	}

	return false
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

	if cond == nil {
		if contains(corev1.ConditionUnknown) {
			return true
		} else {
			return false
		}
	}

	if contains(cond.Status) {
		return true
	} else {
		return false
	}
}

func matchesLatestReason(image v1alpha1.Image, values []string) bool {
	for _, v := range values {
		if strings.ToLower(image.Status.LatestBuildReason) == strings.ToLower(v) {
			return true
		}
	}

	return false
}
