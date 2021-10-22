package config

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s "k8s.io/client-go/kubernetes"
)

const (
	kpConfigMapName                     = "kp-config"
	kpConfigNamespace                   = "kpack"
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

	return sanitize(c.defaultRepository), nil
}

func (c KpConfig) ServiceAccount() corev1.ObjectReference {
	if c.serviceAccount.Name == "" {
		return corev1.ObjectReference{Name: "default", Namespace: kpConfigNamespace}
	}

	return c.serviceAccount
}

type KpConfigProvider struct {
	client k8s.Interface
}

func NewKpConfigProvider(client k8s.Interface) KpConfigProvider {
	return KpConfigProvider{client: client}
}

func (d KpConfigProvider) GetKpConfig(ctx context.Context) KpConfig {
	kpConfig, err := d.getKpConfigMap(ctx)
	if err != nil {
		return KpConfig{}
	}

	repo, ok := kpConfig.Data[defaultRepositoryKey]
	if !ok {
		repo = kpConfig.Data[canonicalRepositoryKey]
	}

	serviceAccountName, ok := kpConfig.Data[defaultServiceAccountNameKey]
	if !ok {
		serviceAccountName = kpConfig.Data[canonicalServiceAccountNameKey]
	}

	serviceAccountNamespace, ok := kpConfig.Data[defaultServiceAccountNamespaceKey]
	if !ok {
		serviceAccountNamespace, ok = kpConfig.Data[canonicalServiceAccountNamespaceKey]
		if !ok {
			serviceAccountNamespace = kpConfigNamespace
		}
	}

	return KpConfig{
		defaultRepository: repo,
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
		return d.createKpConfigMap(ctx, map[string]string{
			defaultRepositoryKey:   defaultRepository,
			canonicalRepositoryKey: defaultRepository,
		})
	}

	return d.updateDefaultRepository(ctx, existingKpConfig, defaultRepository)
}

func (d KpConfigProvider) SetDefaultServiceAccount(ctx context.Context, serviceAccount corev1.ObjectReference) error {
	existingKpConfig, err := d.getKpConfigMap(ctx)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		return d.createKpConfigMap(ctx, map[string]string{
			defaultServiceAccountNameKey:        serviceAccount.Name,
			defaultServiceAccountNamespaceKey:   serviceAccount.Namespace,
			canonicalServiceAccountNameKey:      serviceAccount.Name,
			canonicalServiceAccountNamespaceKey: serviceAccount.Namespace,
		})
	}

	return d.updateDefaultServiceAccount(ctx, existingKpConfig, serviceAccount)
}

func (d KpConfigProvider) getKpConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	return d.client.CoreV1().ConfigMaps(kpConfigNamespace).Get(ctx, kpConfigMapName, metav1.GetOptions{})
}

func (d KpConfigProvider) createKpConfigMap(ctx context.Context, data map[string]string) error {
	kpConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kpConfigMapName,
			Namespace: kpConfigNamespace,
		},
		Data: data,
	}
	_, err := d.client.CoreV1().ConfigMaps(kpConfigNamespace).Create(ctx, kpConfig, metav1.CreateOptions{})
	return err
}

func (d KpConfigProvider) updateDefaultRepository(ctx context.Context, existingConfig *corev1.ConfigMap, repo string) error {
	updatedConfig := existingConfig.DeepCopy()

	updatedConfig.Data[defaultRepositoryKey] = repo
	updatedConfig.Data[canonicalRepositoryKey] = repo

	_, err := d.client.CoreV1().ConfigMaps(kpConfigNamespace).Update(ctx, updatedConfig, metav1.UpdateOptions{})
	return err
}

func (d KpConfigProvider) updateDefaultServiceAccount(ctx context.Context, existingConfig *corev1.ConfigMap, sa corev1.ObjectReference) error {
	updatedConfig := existingConfig.DeepCopy()

	updatedConfig.Data[defaultServiceAccountNameKey] = sa.Name
	updatedConfig.Data[canonicalServiceAccountNameKey] = sa.Name
	updatedConfig.Data[defaultServiceAccountNamespaceKey] = sa.Namespace
	updatedConfig.Data[canonicalServiceAccountNamespaceKey] = sa.Namespace

	_, err := d.client.CoreV1().ConfigMaps(kpConfigNamespace).Update(ctx, updatedConfig, metav1.UpdateOptions{})
	return err
}

func sanitize(r string) string {
	return strings.TrimSuffix(r, "/")
}
