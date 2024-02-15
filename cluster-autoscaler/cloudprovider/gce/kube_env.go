package gce

import (
	"fmt"

	gce "google.golang.org/api/compute/v1"
	"sigs.k8s.io/yaml"
)

const autoscalerVars = "AUTOSCALER_ENV_VARS"

// KubeEnv stores kube-env information from InstanceTemplate
type KubeEnv map[string]string

// ExtractKubeEnv extracts kube-env from InstanceTemplate
func ExtractKubeEnv(template *gce.InstanceTemplate) (KubeEnv, error) {
	if template.Properties.Metadata == nil {
		return nil, fmt.Errorf("instance template %s has no metadata", template.Name)
	}
	for _, item := range template.Properties.Metadata.Items {
		if item.Key == "kube-env" {
			if item.Value == nil {
				return nil, fmt.Errorf("no kube-env content in metadata")
			}
			kubeEnv := make(KubeEnv)
			err := yaml.Unmarshal([]byte(*item.Value), &kubeEnv)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling kubeEnv: %v", err)
			}
			return kubeEnv, nil
		}
	}
	return nil, nil
}
