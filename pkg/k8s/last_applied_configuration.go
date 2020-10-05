package k8s

import (
	"encoding/json"
)

type Annotatable interface {
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
}

const (
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func SetLastAppliedCfg(obj Annotatable) error {
	cfg, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	a := obj.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}

	a[kubectlLastAppliedConfig] = string(cfg)
	obj.SetAnnotations(a)

	return nil
}
