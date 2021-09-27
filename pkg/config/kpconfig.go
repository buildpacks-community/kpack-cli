package config

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s "k8s.io/client-go/kubernetes"
)

const (
	kpConfigNamespace                   = "kpack"
	kpConfigMapName                     = "kp-config"
	defaultRepositoryKey                = "default.repository"
	defaultServiceAccountNameKey        = "default.repository.serviceaccount"
	defaultServiceAccountNamespaceKey   = "default.repository.serviceaccount.namespace"
	canonicalRepositoryKey              = "canonical.repository"                          // historical key
	canonicalServiceAccountNameKey      = "canonical.repository.serviceaccount"           // historical key
	canonicalServiceAccountNamespaceKey = "canonical.repository.serviceaccount.namespace" // historical key
)

type KpConfig struct {
	defaultRepository string
	serviceAccount    corev1.ObjectReference
}

func NewKpConfig(defaultRepository string, serviceAccount corev1.ObjectReference) KpConfig {
	return KpConfig{
		defaultRepository: defaultRepository,
		serviceAccount:    serviceAccount,
	}
}

func (c KpConfig) DefaultRepository() (string, error) {
	if c.defaultRepository == "" {
		return "", errors.New("failed to get default repository: use \"kp config default-repository\" to set")
	}

	return c.defaultRepository, nil
}

func (c KpConfig) ServiceAccount() corev1.ObjectReference {
	if c.serviceAccount.Name == "" {
		return corev1.ObjectReference{Name: "default", Namespace: kpConfigNamespace}
	}

	return c.serviceAccount
}

type KpConfigProvider struct {
	client                             k8s.Interface
	repoKey, saNameKey, saNamespaceKey string
}

func NewKpConfigProvider(ctx context.Context, client k8s.Interface) (KpConfigProvider, error) {
	kpConfigMap, err := client.CoreV1().ConfigMaps(kpConfigNamespace).Get(ctx, kpConfigMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return KpConfigProvider{
			client:         client,
			repoKey:        defaultRepositoryKey,
			saNameKey:      defaultServiceAccountNameKey,
			saNamespaceKey: defaultServiceAccountNamespaceKey,
		}, nil
	} else if err != nil {
		return KpConfigProvider{}, err
	}

	repoKey := defaultRepositoryKey
	if _, ok := kpConfigMap.Data[canonicalRepositoryKey]; ok {
		repoKey = canonicalRepositoryKey
	}

	saNameKey := defaultServiceAccountNameKey
	if _, ok := kpConfigMap.Data[canonicalServiceAccountNameKey]; ok {
		saNameKey = canonicalServiceAccountNameKey
	}

	saNamespaceKey := defaultServiceAccountNamespaceKey
	if _, ok := kpConfigMap.Data[canonicalServiceAccountNamespaceKey]; ok {
		saNamespaceKey = canonicalServiceAccountNamespaceKey
	}

	return KpConfigProvider{
		client:         client,
		repoKey:        repoKey,
		saNameKey:      saNameKey,
		saNamespaceKey: saNamespaceKey,
	}, nil
}

func (d KpConfigProvider) GetKpConfig(ctx context.Context) KpConfig {
	kpConfig, err := d.getKpConfigMap(ctx)
	if err != nil {
		return KpConfig{}
	}

	serviceAccountName := kpConfig.Data[d.saNameKey]
	serviceAccountNamespace, ok := kpConfig.Data[d.saNamespaceKey]
	if !ok {
		serviceAccountNamespace = kpConfigNamespace
	}

	return KpConfig{
		defaultRepository: kpConfig.Data[d.repoKey],
		serviceAccount: corev1.ObjectReference{
			Name:      serviceAccountName,
			Namespace: serviceAccountNamespace,
		},
	}
}

func (d KpConfigProvider) SetDefaultRepository(ctx context.Context, defaultRepository string) error {
	existingKpConfig, err := d.getKpConfigMap(ctx)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		return d.createKpConfigMap(ctx, KpConfig{
			defaultRepository: defaultRepository,
		})
	}

	return d.updateKpConfigMap(ctx, existingKpConfig, KpConfig{defaultRepository: defaultRepository})
}

func (d KpConfigProvider) SetDefaultServiceAccount(ctx context.Context, serviceAccount corev1.ObjectReference) error {
	existingKpConfig, err := d.getKpConfigMap(ctx)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		return d.createKpConfigMap(ctx, KpConfig{
			serviceAccount: serviceAccount,
		})
	}

	return d.updateKpConfigMap(ctx, existingKpConfig, KpConfig{serviceAccount: serviceAccount})
}

func (d KpConfigProvider) getKpConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	return d.client.CoreV1().ConfigMaps(kpConfigNamespace).Get(ctx, kpConfigMapName, metav1.GetOptions{})
}

func (d KpConfigProvider) configMapFromKpConfig(config KpConfig) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kpConfigMapName,
			Namespace: kpConfigNamespace,
		},
		Data: map[string]string{
			d.repoKey:        config.defaultRepository,
			d.saNameKey:      config.serviceAccount.Name,
			d.saNamespaceKey: config.serviceAccount.Namespace,
		},
	}
}

func (d KpConfigProvider) createKpConfigMap(ctx context.Context, config KpConfig) error {
	_, err := d.client.CoreV1().ConfigMaps(kpConfigNamespace).Create(ctx, d.configMapFromKpConfig(config), metav1.CreateOptions{})
	return err
}

func (d KpConfigProvider) updateKpConfigMap(ctx context.Context, existingConfig *corev1.ConfigMap, newConfig KpConfig) error {
	updatedConfig := existingConfig.DeepCopy()

	if newConfig.defaultRepository != "" {
		updatedConfig.Data[d.repoKey] = newConfig.defaultRepository
	}

	if newConfig.serviceAccount.Name != "" {
		updatedConfig.Data[d.saNameKey] = newConfig.serviceAccount.Name
	}

	if newConfig.serviceAccount.Namespace != "" {
		updatedConfig.Data[d.saNamespaceKey] = newConfig.serviceAccount.Namespace
	}

	_, err := d.client.CoreV1().ConfigMaps(kpConfigNamespace).Update(ctx, updatedConfig, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
