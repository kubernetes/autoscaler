/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kwok

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	kubeclient "k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
)

const (
	defaultConfigName = "kwok-provider-config"
	configKey         = "config"
)

// based on https://github.com/kubernetes/kubernetes/pull/63707/files
func getCurrentNamespace() string {
	currentNamespace := os.Getenv("POD_NAMESPACE")
	if strings.TrimSpace(currentNamespace) == "" {
		klog.Info("env variable 'POD_NAMESPACE' is empty")
		klog.Info("trying to read current namespace from serviceaccount")
		// Fall back to the namespace associated with the service account token, if available
		if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
				currentNamespace = ns
			} else {
				klog.Fatal("couldn't get current namespace from serviceaccount")
			}
		} else {
			klog.Fatal("couldn't read serviceaccount to get current namespace")
		}

	}

	klog.Infof("got current pod namespace '%s'", currentNamespace)

	return currentNamespace
}

func getConfigMapName() string {
	configMapName := os.Getenv("KWOK_PROVIDER_CONFIGMAP")
	if strings.TrimSpace(configMapName) == "" {
		klog.Infof("env variable 'KWOK_PROVIDER_CONFIGMAP' is empty (defaulting to '%s')", defaultConfigName)
		configMapName = defaultConfigName
	}

	return configMapName
}

// LoadConfigFile loads kwok provider config from k8s configmap
func LoadConfigFile(kubeClient kubeclient.Interface) (*KwokProviderConfig, error) {
	configMapName := getConfigMapName()

	currentNamespace := getCurrentNamespace()

	c, err := kubeClient.CoreV1().ConfigMaps(currentNamespace).Get(context.Background(), configMapName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap '%s': %v", configMapName, err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(c.Data[configKey]), 4096)
	kwokConfig := KwokProviderConfig{}
	if err := decoder.Decode(&kwokConfig); err != nil {
		return nil, fmt.Errorf("failed to decode kwok config: %v", err)
	}

	if kwokConfig.status == nil {
		kwokConfig.status = &GroupingConfig{}
	}

	switch kwokConfig.ReadNodesFrom {
	case nodeTemplatesFromConfigMap:

		if kwokConfig.ConfigMap == nil {
			return nil, fmt.Errorf("please specify a value for 'configmap' in kwok config (currently empty or undefined)")
		}
		if strings.TrimSpace(kwokConfig.ConfigMap.Name) == "" {
			return nil, fmt.Errorf("please specify 'configmap.name' in kwok config (currently empty or undefined)")
		}

	case nodeTemplatesFromCluster:
	default:
		return nil, fmt.Errorf("'readNodesFrom' in kwok config is invalid (expected: '%s' or '%s'): %s",
			groupNodesByLabel, groupNodesByAnnotation,
			kwokConfig.ReadNodesFrom)
	}

	if kwokConfig.Nodegroups == nil {
		return nil, fmt.Errorf("please specify a value for 'nodegroups' in kwok config (currently empty or undefined)")
	}

	if strings.TrimSpace(kwokConfig.Nodegroups.FromNodeLabelKey) == "" &&
		strings.TrimSpace(kwokConfig.Nodegroups.FromNodeAnnotationKey) == "" {
		return nil, fmt.Errorf("please specify either 'nodegroups.fromNodeLabelKey' or 'nodegroups.fromNodeAnnotationKey' in kwok provider config (currently empty or undefined)")
	}
	if strings.TrimSpace(kwokConfig.Nodegroups.FromNodeLabelKey) != "" &&
		strings.TrimSpace(kwokConfig.Nodegroups.FromNodeAnnotationKey) != "" {
		return nil, fmt.Errorf("please specify either 'nodegroups.fromNodeLabelKey' or 'nodegroups.fromNodeAnnotationKey' in kwok provider config (you can't use both)")
	}

	if strings.TrimSpace(kwokConfig.Nodegroups.FromNodeLabelKey) != "" {
		kwokConfig.status.groupNodesBy = groupNodesByLabel
		kwokConfig.status.key = kwokConfig.Nodegroups.FromNodeLabelKey
	} else {
		kwokConfig.status.groupNodesBy = groupNodesByAnnotation
		kwokConfig.status.key = kwokConfig.Nodegroups.FromNodeAnnotationKey
	}

	if kwokConfig.Nodes == nil {
		kwokConfig.Nodes = &NodeConfig{}
	} else {

		if kwokConfig.Nodes.GPUConfig == nil {
			klog.Warningf("nodes.gpuConfig is empty or undefined")
		} else {
			if kwokConfig.Nodes.GPUConfig.GPULabelKey != "" &&
				kwokConfig.Nodes.GPUConfig.AvailableGPUTypes != nil {
				kwokConfig.status.availableGPUTypes = kwokConfig.Nodes.GPUConfig.AvailableGPUTypes
				kwokConfig.status.gpuLabel = kwokConfig.Nodes.GPUConfig.GPULabelKey
			} else {
				return nil, errors.New("nodes.gpuConfig.gpuLabelKey or file.nodes.gpuConfig.availableGPUTypes is empty")
			}
		}

	}

	if kwokConfig.Kwok == nil {
		kwokConfig.Kwok = &KwokConfig{}
	}

	return &kwokConfig, nil
}
