package config

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

const (
	kpNamespace                       = "kpack"
	kpConfigMapName                   = "kp-config"
	defaultRepositoryKey              = "default.repository"
	defaultServiceAccountNameKey      = "default.repository.serviceaccount"
	defaultServiceAccountNamespaceKey = "default.repository.serviceaccount.namespace"
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
		return corev1.ObjectReference{Name: "default", Namespace: kpNamespace}
	}

	return c.serviceAccount
}

type KpConfigProvider struct {
	cs k8s.ClientSet
}

func NewKpConfigProvider(cs k8s.ClientSet) KpConfigProvider {
	return KpConfigProvider{cs: cs}
}

func (d KpConfigProvider) GetKpConfig(ctx context.Context) KpConfig {
	kpConfig, err := d.getKpConfigMap(ctx)
	if err != nil {
		return KpConfig{}
	}

	serviceAccountName := kpConfig.Data[defaultServiceAccountNameKey]
	serviceAccountNamespace, ok := kpConfig.Data[defaultServiceAccountNamespaceKey]
	if !ok && serviceAccountName != "" {
		serviceAccountNamespace = kpNamespace
	}

	return KpConfig{
		defaultRepository: kpConfig.Data[defaultRepositoryKey],
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
		return createKpConfigMap(ctx, d.cs.K8sClient, KpConfig{
			defaultRepository: defaultRepository,
		})
	}

	return updateKpConfigMap(ctx, d.cs.K8sClient, existingKpConfig, KpConfig{defaultRepository: defaultRepository})
}

func (d KpConfigProvider) SetDefaultServiceAccount(ctx context.Context, serviceAccount corev1.ObjectReference) error {
	existingKpConfig, err := d.getKpConfigMap(ctx)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		return createKpConfigMap(ctx, d.cs.K8sClient, KpConfig{
			serviceAccount: serviceAccount,
		})
	}

	return updateKpConfigMap(ctx, d.cs.K8sClient, existingKpConfig, KpConfig{serviceAccount: serviceAccount})
}

func (d KpConfigProvider) getKpConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	return d.cs.K8sClient.CoreV1().ConfigMaps(kpNamespace).Get(ctx, kpConfigMapName, metav1.GetOptions{})
}

func configMapFromKpConfig(config KpConfig) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kpConfigMapName,
			Namespace: kpNamespace,
		},
		Data: map[string]string{
			defaultRepositoryKey:              config.defaultRepository,
			defaultServiceAccountNameKey:      config.serviceAccount.Name,
			defaultServiceAccountNamespaceKey: config.serviceAccount.Namespace,
		},
	}
}

func createKpConfigMap(ctx context.Context, client kubernetes.Interface, config KpConfig) error {
	_, err := client.CoreV1().ConfigMaps(kpNamespace).Create(ctx, configMapFromKpConfig(config), metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func updateKpConfigMap(ctx context.Context, client kubernetes.Interface, existingConfig *corev1.ConfigMap, newConfig KpConfig) error {
	updatedConfig := existingConfig.DeepCopy()

	if newConfig.defaultRepository != "" {
		updatedConfig.Data[defaultRepositoryKey] = newConfig.defaultRepository
	}

	if newConfig.serviceAccount.Name != "" {
		updatedConfig.Data[defaultServiceAccountNameKey] = newConfig.serviceAccount.Name
	}

	if newConfig.serviceAccount.Namespace != "" {
		updatedConfig.Data[defaultServiceAccountNamespaceKey] = newConfig.serviceAccount.Namespace
	}

	_, err := client.CoreV1().ConfigMaps(kpNamespace).Update(ctx, updatedConfig, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
