package build

import (
	"strings"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

func sortBuilds(builds []v1alpha1.Build) func(i int, j int) bool {
	return func(i, j int) bool {
		return builds[j].ObjectMeta.CreationTimestamp.After(builds[i].ObjectMeta.CreationTimestamp.Time)
	}
}

func getStatus(b v1alpha1.Build) string {
	cond := b.Status.GetCondition(corev1alpha1.ConditionSucceeded)
	switch {
	case cond.IsTrue():
		return "SUCCESS"
	case cond.IsFalse():
		return "FAILURE"
	case cond.IsUnknown():
		return "BUILDING"
	default:
		return "UNKNOWN"
	}
}

func getStarted(b v1alpha1.Build) string {
	return b.CreationTimestamp.Time.Format("2006-01-02 15:04:05")
}

func getFinished(b v1alpha1.Build) string {
	if b.IsRunning() {
		return ""
	}
	return b.Status.GetCondition(corev1alpha1.ConditionSucceeded).LastTransitionTime.Inner.Format("2006-01-02 15:04:05")
}

func getTruncatedReason(b v1alpha1.Build) string {
	r := getReasons(b)

	if len(r) == 0 {
		return "UNKNOWN"
	}

	if len(r) == 1 {
		return r[0]
	}

	return mostImportantReason(r) + "+"
}

func getReasons(b v1alpha1.Build) []string {
	s := strings.Split(b.Annotations[v1alpha1.BuildReasonAnnotation], ",")
	if len(s) == 1 && s[0] == "" {
		return nil
	}
	return s
}

func mostImportantReason(r []string) string {
	if contains(r, "CONFIG") {
		return "CONFIG"
	} else if contains(r, "COMMIT") {
		return "COMMIT"
	} else if contains(r, "BUILDPACK") {
		return "BUILDPACK"
	}

	return r[0]
}

func contains(reasons []string, value string) bool {
	for _, v := range reasons {
		if v == value {
			return true
		}
	}
	return false
}
