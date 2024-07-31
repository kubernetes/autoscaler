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

package azure

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	autoDiscovererTypeLabel       = "label"
	vmssAutoDiscovererKeyMinNodes = "min"
	vmssAutoDiscovererKeyMaxNodes = "max"
)

// A labelAutoDiscoveryConfig specifies how to auto-discover Azure node groups.
type labelAutoDiscoveryConfig struct {
	// Key-values to match on.
	Selector map[string]string
	// MinSize specifies the minimum size for all VMSSs that match Selector.
	MinSize *int
	// MazSize specifies the maximum size for all VMSSs that match Selector.
	MaxSize *int
}

type autoDiscoveryConfigSizes struct {
	Min int
	Max int
}

// ParseLabelAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for node group autodiscovery.
func ParseLabelAutoDiscoverySpecs(o cloudprovider.NodeGroupDiscoveryOptions) ([]labelAutoDiscoveryConfig, error) {
	cfgs := make([]labelAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseLabelAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

// parseLabelAutoDiscoverySpec parses a single spec and returns the corresponding node group spec.
func parseLabelAutoDiscoverySpec(spec string) (labelAutoDiscoveryConfig, error) {
	cfg := labelAutoDiscoveryConfig{
		Selector: make(map[string]string),
	}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("spec \"%s\" should be discoverer:key=value,key=value", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeLabel {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}

	for _, arg := range strings.Split(tokens[1], ",") {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return cfg, fmt.Errorf("invalid key=value pair %s", kv)
		}
		k, v := kv[0], kv[1]
		if k == "" || v == "" {
			return cfg, fmt.Errorf("empty value not allowed in key=value tag pairs")
		}

		switch k {
		case vmssAutoDiscovererKeyMinNodes:
			minSize, err := strconv.Atoi(v)
			if err != nil || minSize < 0 {
				return cfg, fmt.Errorf("invalid minimum nodes: %s", v)
			}
			cfg.MinSize = &minSize
		case vmssAutoDiscovererKeyMaxNodes:
			maxSize, err := strconv.Atoi(v)
			if err != nil || maxSize < 0 {
				return cfg, fmt.Errorf("invalid maximum nodes: %s", v)
			}
			cfg.MaxSize = &maxSize
		default:
			cfg.Selector[k] = v
		}
	}
	if cfg.MaxSize != nil && cfg.MinSize != nil && *cfg.MaxSize < *cfg.MinSize {
		return cfg, fmt.Errorf("maximum size %d must be greater than or equal to minimum size %d", *cfg.MaxSize, *cfg.MinSize)
	}
	return cfg, nil
}

// returns an autoDiscoveryConfigSizes struct if the VMSS's tags match the autodiscovery configs
// if the VMSS's tags do not match then return nil
// if there are multiple min/max sizes defined, return the highest min value and the lowest max value
func matchDiscoveryConfig(labels map[string]*string, configs []labelAutoDiscoveryConfig) *autoDiscoveryConfigSizes {
	if len(configs) == 0 {
		return nil
	}
	minSize := -1
	maxSize := -1

	for _, c := range configs {
		if len(c.Selector) == 0 {
			return nil
		}

		for k, v := range c.Selector {
			value, ok := labels[k]
			if !ok {
				return nil
			}

			if v != "" {
				if value == nil || *value != v {
					return nil
				}
			}
		}
		if c.MinSize != nil && minSize < *c.MinSize {
			minSize = *c.MinSize
		}
		if c.MaxSize != nil && (maxSize == -1 || maxSize > *c.MaxSize) {
			maxSize = *c.MaxSize
		}
	}

	return &autoDiscoveryConfigSizes{
		Min: minSize,
		Max: maxSize,
	}
}
