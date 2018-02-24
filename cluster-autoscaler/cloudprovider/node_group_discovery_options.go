/*
Copyright 2016 The Kubernetes Authors.

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

package cloudprovider

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	kubeclient "k8s.io/client-go/kubernetes"
)

const (
	autoDiscovererTypeMIG   = "mig"
	autoDiscovererTypeASG   = "asg"
	autoDiscovererTypeLabel = "label"

	migAutoDiscovererKeyPrefix   = "namePrefix"
	migAutoDiscovererKeyMinNodes = "min"
	migAutoDiscovererKeyMaxNodes = "max"

	asgAutoDiscovererKeyTag = "tag"
)

var validMIGAutoDiscovererKeys = strings.Join([]string{
	migAutoDiscovererKeyPrefix,
	migAutoDiscovererKeyMinNodes,
	migAutoDiscovererKeyMaxNodes,
}, ", ")

// NodeGroupDiscoveryOptions contains various options to configure how a cloud provider discovers node groups
type NodeGroupDiscoveryOptions struct {
	// NodeGroupSpecs is specified to statically discover node groups listed in it
	NodeGroupSpecs []string
	// NodeGroupAutoDiscoverySpec is specified for automatically discovering node groups according to the specs
	NodeGroupAutoDiscoverySpecs []string
	// KubeClient is used for cloud provider to fetch node list.
	KubeClient kubeclient.Interface
}

// StaticDiscoverySpecified returns true only when there are 1 or more --nodes flags specified
func (o NodeGroupDiscoveryOptions) StaticDiscoverySpecified() bool {
	return len(o.NodeGroupSpecs) > 0
}

// AutoDiscoverySpecified returns true only when there are 1 or more --node-group-auto-discovery flags specified
func (o NodeGroupDiscoveryOptions) AutoDiscoverySpecified() bool {
	return len(o.NodeGroupAutoDiscoverySpecs) > 0
}

// DiscoverySpecified returns true when at least one of the --nodes or
// --node-group-auto-discovery flags specified.
func (o NodeGroupDiscoveryOptions) DiscoverySpecified() bool {
	return o.StaticDiscoverySpecified() || o.AutoDiscoverySpecified()
}

// ParseMIGAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for MIG autodiscovery.
func (o NodeGroupDiscoveryOptions) ParseMIGAutoDiscoverySpecs() ([]MIGAutoDiscoveryConfig, error) {
	cfgs := make([]MIGAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseMIGAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

// ParseASGAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for ASG autodiscovery.
func (o NodeGroupDiscoveryOptions) ParseASGAutoDiscoverySpecs() ([]ASGAutoDiscoveryConfig, error) {
	cfgs := make([]ASGAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseASGAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

// ParseLabelAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for ASG autodiscovery.
func (o NodeGroupDiscoveryOptions) ParseLabelAutoDiscoverySpecs() ([]LabelAutoDiscoveryConfig, error) {
	cfgs := make([]LabelAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseLabelAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

// A MIGAutoDiscoveryConfig specifies how to autodiscover GCE MIGs.
type MIGAutoDiscoveryConfig struct {
	// Re is a regexp passed using the eq filter to the GCE list API.
	Re *regexp.Regexp
	// MinSize specifies the minimum size for all MIGs that match Re.
	MinSize int
	// MaxSize specifies the maximum size for all MIGs that match Re.
	MaxSize int
}

func parseMIGAutoDiscoverySpec(spec string) (MIGAutoDiscoveryConfig, error) {
	cfg := MIGAutoDiscoveryConfig{}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("spec \"%s\" should be discoverer:key=value,key=value", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeMIG {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}

	for _, arg := range strings.Split(tokens[1], ",") {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return cfg, fmt.Errorf("invalid key=value pair %s", kv)
		}
		k, v := kv[0], kv[1]

		var err error
		switch k {
		case migAutoDiscovererKeyPrefix:
			if cfg.Re, err = regexp.Compile(fmt.Sprintf("^%s.+", v)); err != nil {
				return cfg, fmt.Errorf("invalid instance group name prefix \"%s\" - \"^%s.+\" must be a valid RE2 regexp", v, v)
			}
		case migAutoDiscovererKeyMinNodes:
			if cfg.MinSize, err = strconv.Atoi(v); err != nil {
				return cfg, fmt.Errorf("invalid minimum nodes: %s", v)
			}
		case migAutoDiscovererKeyMaxNodes:
			if cfg.MaxSize, err = strconv.Atoi(v); err != nil {
				return cfg, fmt.Errorf("invalid maximum nodes: %s", v)
			}
		default:
			return cfg, fmt.Errorf("unsupported key \"%s\" is specified for discoverer \"%s\". Supported keys are \"%s\"", k, discoverer, validMIGAutoDiscovererKeys)
		}
	}
	if cfg.Re == nil || cfg.Re.String() == "^.+" {
		return cfg, errors.New("empty instance group name prefix supplied")
	}
	if cfg.MinSize > cfg.MaxSize {
		return cfg, fmt.Errorf("minimum size %d is greater than maximum size %d", cfg.MinSize, cfg.MaxSize)
	}
	if cfg.MaxSize < 1 {
		return cfg, fmt.Errorf("maximum size %d must be at least 1", cfg.MaxSize)
	}
	return cfg, nil
}

// An ASGAutoDiscoveryConfig specifies how to autodiscover AWS ASGs.
type ASGAutoDiscoveryConfig struct {
	// TagKeys to match on.
	// Any ASG with all of the provided tag keys wMIGAutoDiscoveryConfigill be autoscaled.
	TagKeys []string
}

func parseASGAutoDiscoverySpec(spec string) (ASGAutoDiscoveryConfig, error) {
	cfg := ASGAutoDiscoveryConfig{}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("Invalid node group auto discovery spec specified via --node-group-auto-discovery: %s", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeASG {
		return cfg, fmt.Errorf("Unsupported discoverer specified: %s", discoverer)
	}
	param := tokens[1]
	kv := strings.Split(param, "=")
	if len(kv) != 2 {
		return cfg, fmt.Errorf("invalid key=value pair %s", kv)
	}
	k, v := kv[0], kv[1]
	if k != asgAutoDiscovererKeyTag {
		return cfg, fmt.Errorf("Unsupported parameter key \"%s\" is specified for discoverer \"%s\". The only supported key is \"%s\"", k, discoverer, asgAutoDiscovererKeyTag)
	}
	if v == "" {
		return cfg, errors.New("tag value not supplied")
	}
	cfg.TagKeys = strings.Split(v, ",")
	if len(cfg.TagKeys) == 0 {
		return cfg, fmt.Errorf("Invalid ASG tag for auto discovery specified: ASG tag must not be empty")
	}
	return cfg, nil
}

// A LabelAutoDiscoveryConfig specifies how to autodiscover Azure scale sets.
type LabelAutoDiscoveryConfig struct {
	// Key-values to match on.
	Selector map[string]string
}

func parseLabelAutoDiscoverySpec(spec string) (LabelAutoDiscoveryConfig, error) {
	cfg := LabelAutoDiscoveryConfig{
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
		cfg.Selector[k] = v
	}

	return cfg, nil
}
