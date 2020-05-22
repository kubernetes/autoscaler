/*
Copyright 2019 The Kubernetes Authors.

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

package magnum

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	// Constants in the node group autodiscovery configuration string.
	autoDiscovererTypeMagnum    = "magnum"
	magnumAutoDiscovererKeyRole = "role"
)

type magnumAutoDiscoveryConfig struct {
	Roles []string
}

func parseMagnumAutoDiscoverySpecs(o cloudprovider.NodeGroupDiscoveryOptions) ([]magnumAutoDiscoveryConfig, error) {
	var cfgs []magnumAutoDiscoveryConfig
	for _, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfg, err := parseMagnumAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs, nil
}

// parseMagnumAutoDiscoverySpec takes a string given via --node-group-auto-discovery
// and parses it into an auto discovery config.
//
// The spec format is:
// magnum:role=<role>[,<role2>]
func parseMagnumAutoDiscoverySpec(spec string) (magnumAutoDiscoveryConfig, error) {
	cfg := magnumAutoDiscoveryConfig{}

	// Split the spec into two parts, the discoverer (magnum)
	// and the discovery parameter (role=value).
	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("invalid node group auto discovery spec specified via --node-group-auto-discovery: %s", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeMagnum {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}

	// Split the discovery parameter into a key value pair.
	param := tokens[1]
	kv := strings.SplitN(param, "=", 2)
	if len(kv) != 2 {
		return cfg, fmt.Errorf("invalid discovery key=value pair %s", kv)
	}

	k, v := kv[0], kv[1]
	if k != magnumAutoDiscovererKeyRole {
		return cfg, fmt.Errorf("unsupported parameter key %q is specified for discoverer %q. The only supported key is %q", k, discoverer, magnumAutoDiscovererKeyRole)
	}

	if v == "" {
		return cfg, errors.New("role value not supplied")
	}

	// Allow specifying multiple roles in a single spec, comma separated.
	roles := strings.Split(v, ",")
	if len(roles) == 0 {
		return cfg, errors.New("no roles specified")
	}

	// Check that all roles are valid.
	for _, r := range roles {
		if len(r) == 0 {
			return cfg, fmt.Errorf("invalid role for auto discovery specified: role must not be empty")
		}
	}

	cfg.Roles = roles

	return cfg, nil
}
