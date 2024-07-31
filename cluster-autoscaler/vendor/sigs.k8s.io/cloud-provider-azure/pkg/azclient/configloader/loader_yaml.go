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

package configloader

import (
	"bytes"
	"context"

	"sigs.k8s.io/yaml"
)

// yamlByteLoader is a FactoryConfigLoader that loads a YAML file from a byte array.
type yamlByteLoader[Type any] struct {
	content []byte
	configLoader[Type]
}

// Load loads the YAML file from the byte array and returns the client factory config.
func (s *yamlByteLoader[Type]) Load(ctx context.Context) (*Type, error) {
	if s.configLoader == nil {
		s.configLoader = newEmptyLoader[Type](nil)
	}
	config, err := s.configLoader.Load(ctx)
	if err != nil {
		return nil, err
	}
	s.content = bytes.TrimSpace(s.content)
	if err := yaml.Unmarshal(bytes.TrimSpace(s.content), config); err != nil {
		return nil, err
	}
	return config, nil
}

// newYamlByteLoader creates a YamlByteLoader with the specified content and loader.
func newYamlByteLoader[Type any](content []byte, loader configLoader[Type]) configLoader[Type] {
	return &yamlByteLoader[Type]{
		content:      content,
		configLoader: loader,
	}
}
