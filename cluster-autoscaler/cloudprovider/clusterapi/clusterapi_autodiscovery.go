/*
Copyright 2020 The Kubernetes Authors.

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

package clusterapi

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

type clusterAPIAutoDiscoveryConfig struct {
	clusterName   string
	namespace     string
	labelSelector labels.Selector
}

func parseAutoDiscoverySpec(spec string) (*clusterAPIAutoDiscoveryConfig, error) {
	cfg := &clusterAPIAutoDiscoveryConfig{
		labelSelector: labels.NewSelector(),
	}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, errors.NewAutoscalerError(errors.ConfigurationError, fmt.Sprintf("spec \"%s\" should be discoverer:key=value,key=value", spec))
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeClusterAPI {
		return cfg, errors.NewAutoscalerError(errors.ConfigurationError, fmt.Sprintf("unsupported discoverer specified: %s", discoverer))
	}

	for _, arg := range strings.Split(tokens[1], ",") {
		if len(arg) == 0 {
			continue
		}
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return cfg, errors.NewAutoscalerError(errors.ConfigurationError, fmt.Sprintf("invalid key=value pair %s", kv))
		}
		k, v := kv[0], kv[1]

		switch k {
		case autoDiscovererClusterNameKey:
			cfg.clusterName = v
		case autoDiscovererNamespaceKey:
			cfg.namespace = v
		default:
			req, err := labels.NewRequirement(k, selection.Equals, []string{v})
			if err != nil {
				return cfg, errors.NewAutoscalerError(errors.ConfigurationError, fmt.Sprintf("failed to create label selector; %v", err))
			}
			cfg.labelSelector = cfg.labelSelector.Add(*req)
		}
	}
	return cfg, nil
}

func parseAutoDiscovery(specs []string) ([]*clusterAPIAutoDiscoveryConfig, error) {
	result := make([]*clusterAPIAutoDiscoveryConfig, 0, len(specs))
	for _, spec := range specs {
		autoDiscoverySpec, err := parseAutoDiscoverySpec(spec)
		if err != nil {
			return result, err
		}
		result = append(result, autoDiscoverySpec)
	}
	return result, nil
}

func allowedByAutoDiscoverySpec(spec *clusterAPIAutoDiscoveryConfig, r *unstructured.Unstructured) bool {
	switch {
	case spec.namespace != "" && spec.namespace != r.GetNamespace():
		return false
	case spec.clusterName != "" && spec.clusterName != clusterNameFromResource(r):
		return false
	case !spec.labelSelector.Matches(labels.Set(r.GetLabels())):
		return false
	default:
		return true
	}
}
