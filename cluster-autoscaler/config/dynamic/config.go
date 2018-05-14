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

package dynamic

import (
	"fmt"
)

// Config which represents not static but dynamic configuration of cluster-autoscaler which would be updated periodically at runtime
type Config struct {
	Settings
	resourceVersion string
}

// Settings of cluster-autoscaler contained in the latest config, which should be consumed by cluster-autoscaler
type Settings struct {
	NodeGroups []NodeGroupSpec `json:"nodeGroups"`
}

// NewDefaultConfig builds a new config object
func NewDefaultConfig() Config {
	return Config{
		Settings: Settings{
			NodeGroups: []NodeGroupSpec{},
		},
		resourceVersion: "",
	}
}

// NodeGroupSpecStrings returns node group specs represented in the form of `<minSize>:<maxSize>:<name>` to be passed to cloudprovider.
func (c Config) NodeGroupSpecStrings() []string {
	return c.nodeGroupSpecStrings()
}

func (c Config) validate() error {
	for _, g := range c.NodeGroups {
		if err := g.Validate(); err != nil {
			return fmt.Errorf("invalid node group: %v", err)
		}
	}
	return nil
}

func (s Settings) nodeGroupSpecStrings() []string {
	result := []string{}

	for _, spec := range s.NodeGroups {
		result = append(result, spec.String())
	}

	return result
}
