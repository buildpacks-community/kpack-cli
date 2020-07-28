package k8s

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigHelper interface {
	GetCanonicalRepository() (string, error)
	GetCanonicalServiceAccount() (string, error)
}

const (
	kpNamespace                = "kpack"
	kpConfigMapName            = "kp-config"
	canonicalRepositoryKey     = "canonical.repository"
	canonicalServiceAccountKey = "canonical.repository.serviceaccount"
)

type defaultConfigHelper struct {
	cs ClientSet
}

func DefaultConfigHelper(cs ClientSet) ConfigHelper {
	return defaultConfigHelper{cs: cs}
}

func (d defaultConfigHelper) GetCanonicalRepository() (string, error) {
	val, err := d.getValue(canonicalRepositoryKey)
	if err != nil {
		return val, errors.Wrapf(err, "failed to get canonical repository")
	}

	if val == "" {
		return "", errors.Errorf("failed to get canonical repository")
	}
	return val, err
}

func (d defaultConfigHelper) GetCanonicalServiceAccount() (string, error) {
	val, err := d.getValue(canonicalServiceAccountKey)
	if err != nil {
		return val, errors.Wrapf(err, "failed to get canonical service account")
	}

	if val == "" {
		return "", errors.Errorf("failed to get canonical service account")
	}
	return val, err
}

func (d defaultConfigHelper) getValue(key string) (string, error) {
	var value string

	kpConfig, err := d.cs.K8sClient.CoreV1().ConfigMaps(kpNamespace).Get(kpConfigMapName, metav1.GetOptions{})
	if err != nil {
		return value, err
	}

	value, ok := kpConfig.Data[key]
	if !ok {
		return value, errors.Errorf("key %q not found in configmap %q", key, kpConfigMapName)
	}

	return value, nil
}
