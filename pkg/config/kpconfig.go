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
	kpNamespace                         = "kpack"
	kpConfigMapName                     = "kp-config"
	canonicalRepositoryKey              = "canonical.repository"
	canonicalServiceAccountNameKey      = "canonical.repository.serviceaccount"
	canonicalServiceAccountNamespaceKey = "canonical.repository.serviceaccount.namespace"
)

type KpConfig struct {
	canonicalRepository string
	serviceAccount      corev1.ObjectReference
}

func NewKpConfig(canonicalRepository string, serviceAccount corev1.ObjectReference) KpConfig {
	return KpConfig{
		canonicalRepository: canonicalRepository,
		serviceAccount:      serviceAccount,
	}
}

func (c KpConfig) CanonicalRepository() (string, error) {
	if c.canonicalRepository == "" {
		return "", errors.New("failed to get canonical repository: use \"kp config canonical-repository\" to set")
	}

	return c.canonicalRepository, nil
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

	serviceAccountName := kpConfig.Data[canonicalServiceAccountNameKey]
	serviceAccountNamespace, ok := kpConfig.Data[canonicalServiceAccountNamespaceKey]
	if !ok && serviceAccountName != "" {
		serviceAccountNamespace = kpNamespace
	}

	return KpConfig{
		canonicalRepository: kpConfig.Data[canonicalRepositoryKey],
		serviceAccount: corev1.ObjectReference{
			Name:      serviceAccountName,
			Namespace: serviceAccountNamespace,
		},
	}
}

func (d KpConfigProvider) SetCanonicalRepository(ctx context.Context, canonicalRepository string) error {
	existingKpConfig, err := d.getKpConfigMap(ctx)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		return createKpConfigMap(ctx, d.cs.K8sClient, KpConfig{
			canonicalRepository: canonicalRepository,
		})
	}

	return updateKpConfigMap(ctx, d.cs.K8sClient, existingKpConfig, KpConfig{canonicalRepository: canonicalRepository})
}

func (d KpConfigProvider) SetCanonicalServiceAccount(ctx context.Context, serviceAccount corev1.ObjectReference) error {
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
			canonicalRepositoryKey:              config.canonicalRepository,
			canonicalServiceAccountNameKey:      config.serviceAccount.Name,
			canonicalServiceAccountNamespaceKey: config.serviceAccount.Namespace,
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

	if newConfig.canonicalRepository != "" {
		updatedConfig.Data[canonicalRepositoryKey] = newConfig.canonicalRepository
	}

	if newConfig.serviceAccount.Name != "" {
		updatedConfig.Data[canonicalServiceAccountNameKey] = newConfig.serviceAccount.Name
	}

	if newConfig.serviceAccount.Namespace != "" {
		updatedConfig.Data[canonicalServiceAccountNamespaceKey] = newConfig.serviceAccount.Namespace
	}

	_, err := client.CoreV1().ConfigMaps(kpNamespace).Update(ctx, updatedConfig, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
