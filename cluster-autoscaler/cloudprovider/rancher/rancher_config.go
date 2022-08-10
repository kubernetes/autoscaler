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

package rancher

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type cloudConfig struct {
	URL               string `yaml:"url"`
	Token             string `yaml:"token"`
	ClusterName       string `yaml:"clusterName"`
	ClusterNamespace  string `yaml:"clusterNamespace"`
	ClusterAPIVersion string `yaml:"clusterAPIVersion"`
}

func newConfig(file string) (*cloudConfig, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read cloud config file: %w", err)
	}

	config := &cloudConfig{}
	if err := yaml.Unmarshal(b, config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file: %w", err)
	}

	return config, nil
}
