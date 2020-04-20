package secret

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

const ManagedSecretAnnotationKey = "build.pivotal.io/managedSecret"

func readManagedSecrets(sa *corev1.ServiceAccount) (map[string]string, error) {
	if sa.Annotations == nil {
		return map[string]string{}, nil
	}

	annotation := sa.Annotations[ManagedSecretAnnotationKey]
	if annotation == "" {
		return map[string]string{}, nil
	}

	managedSecrets := map[string]string{}
	err := json.Unmarshal([]byte(annotation), &managedSecrets)
	if err != nil {
		return map[string]string{}, err
	}

	return managedSecrets, nil
}

func writeManagedSecrets(managedSecrets map[string]string, sa *corev1.ServiceAccount) error {
	if sa.Annotations == nil {
		sa.Annotations = map[string]string{}
	}

	if len(managedSecrets) == 0 {
		delete(sa.Annotations, ManagedSecretAnnotationKey)
	} else {

		buf, err := json.Marshal(managedSecrets)
		if err != nil {
			return err
		}

		sa.Annotations[ManagedSecretAnnotationKey] = string(buf)
	}

	return nil
}
