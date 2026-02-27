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
	"os"
	"reflect"

	"k8s.io/klog/v2"
)

// ConfigFetcher fetches the up-to-date dynamic configuration from the mounted configmap
type ConfigFetcher interface {
	FetchConfigIfUpdated() (*Config, error)
}

type configFetcher struct {
	path                  string
	previousConfiguration Config
}

// ConfigFetcherOptions contains options to customize the ConfigFetcher
type ConfigFetcherOptions struct {
	ConfigPath string
}

// NewConfigFetcher builds a config fetcher
func NewConfigFetcher(options ConfigFetcherOptions) *configFetcher {
	return &configFetcher{
		path:                  options.ConfigPath,
		previousConfiguration: NewDefaultConfig(),
	}
}

// Returns the config if it has changed and nil if it hasn't
func (cf *configFetcher) FetchConfigIfUpdated() (*Config, error) {
	configFile, err := os.Open(cf.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}

	config, err := BuildConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load dynamic config: %v", err)
	}

	if err := configFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close config file: %v", err)
	}

	if reflect.DeepEqual(cf.previousConfiguration, *config) {
		klog.V(4).Infof("config matches previous one - no need to update")
		return nil, nil
	}

	// else config has changed, update
	cf.previousConfiguration = *config
	return config, nil
}

// Returns the config regardless of whether it has changed, also updates the previous configuration
func (cf *configFetcher) ForceFetchConfig() (*Config, error) {
	configFile, err := os.Open(cf.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}

	config, err := BuildConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load dynamic config: %v", err)
	}

	if err := configFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close config file: %v", err)
	}

	cf.previousConfiguration = *config
	return config, nil
}
