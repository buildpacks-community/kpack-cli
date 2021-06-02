package config

import corev1 "k8s.io/api/core/v1"

type KpConfig struct {
	CanonicalRepository string
	ServiceAccount      corev1.ObjectReference
}
