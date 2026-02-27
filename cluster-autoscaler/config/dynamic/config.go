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

package dynamic

import (
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
	"k8s.io/klog/v2"
)

// Config holds the dynamic configuration of autoscaler which can be refreshed at runtime
type Config struct {
	NodeGroups []NodeGroupSpec `json:"nodeGroups" yaml:"nodeGroups"`
}

// NewDefaultConfig returns a default empty config
func NewDefaultConfig() Config {
	return Config{
		NodeGroups: []NodeGroupSpec{},
	}
}

// BuildConfig builds a Config object from the mounted path
func BuildConfig(reader io.Reader) (*Config, error) {
	config, err := umarshalConfig(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode autoscaler config: %v", err)
	}

	klog.V(4).Infof("nodeGroups=%v", config.NodeGroups)

	var modifiedNg []NodeGroupSpec
	for _, spec := range config.NodeGroups {
		localSpec := spec
		if spec.ScaleDownPolicy == "" {
			localSpec.ScaleDownPolicy = deallocate.Delete
		}
		modifiedNg = append(modifiedNg, localSpec)
	}
	config.NodeGroups = modifiedNg
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("error while validating config: %v", err)
	}

	return &config, nil
}

// umarshalConfig decodes the yaml or json reader into a struct
func umarshalConfig(reader io.Reader) (Config, error) {
	config := Config{}
	if err := yaml.NewYAMLOrJSONDecoder(reader, 4096).Decode(&config); err != nil {
		return Config{}, err
	}
	return config, nil
}

// NodeGroupSpecStrings returns node group specs represented in the form of `<minSize>:<maxSize>:<name>:<labels>|<taints>` to be passed to
// the cloudprovider autoscaling options
func (c Config) NodeGroupSpecStrings() []string {
	result := []string{}
	for _, spec := range c.NodeGroups {
		result = append(result, spec.StringWithLabelsAndTaints())
	}
	return result
}

func (c Config) validate() error {
	for _, g := range c.NodeGroups {
		if g.Name == "" {
			return fmt.Errorf("invalid nodeGroup: name must not be blank")
		}
		if g.MaxSize < g.MinSize {
			return fmt.Errorf("invalid nodeGroup: %s, max size must be greater or equal to min size", g.Name)
		}
		if g.ScaleDownPolicy != deallocate.Delete && g.ScaleDownPolicy != deallocate.Deallocate {
			return fmt.Errorf("invalid scaledown policy: %s. Valid values are: Delete, Deallocate", g.ScaleDownPolicy)
		}
	}
	return nil
}
