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

package provider

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/yaml"
)

// The config type for Azure cloud provider secret. Supported values are:
// * file   : The values are read from local cloud-config file.
// * secret : The values from secret would override all configures from local cloud-config file.
// * merge  : The values from secret would override only configurations that are explicitly set in the secret. This is the default value.
type cloudConfigType string

const (
	cloudConfigTypeFile   cloudConfigType = "file"
	cloudConfigTypeSecret cloudConfigType = "secret"
	cloudConfigTypeMerge  cloudConfigType = "merge"
)

// InitializeCloudFromSecret initializes Azure cloud provider from Kubernetes secret.
func (az *Cloud) InitializeCloudFromSecret() error {
	config, err := az.GetConfigFromSecret()
	if err != nil {
		klog.Errorf("Failed to get cloud-config from secret: %v", err)
		return fmt.Errorf("InitializeCloudFromSecret: failed to get cloud config from secret %s/%s: %w", az.SecretNamespace, az.SecretName, err)
	}

	if config == nil {
		// Skip re-initialization if the config is not override.
		return nil
	}

	if err := az.InitializeCloudFromConfig(config, true, true); err != nil {
		klog.Errorf("Failed to initialize Azure cloud provider: %v", err)
		return fmt.Errorf("InitializeCloudFromSecret: failed to initialize Azure cloud provider: %w", err)
	}

	return nil
}

func (az *Cloud) GetConfigFromSecret() (*Config, error) {
	// Read config from file and no override, return nil.
	if az.Config.CloudConfigType == cloudConfigTypeFile {
		return nil, nil
	}

	secret, err := az.KubeClient.CoreV1().Secrets(az.SecretNamespace).Get(context.TODO(), az.SecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", az.SecretNamespace, az.SecretName, err)
	}

	cloudConfigData, ok := secret.Data[az.CloudConfigKey]
	if !ok {
		return nil, fmt.Errorf("cloud-config is not set in the secret (%s/%s)", az.SecretNamespace, az.SecretName)
	}

	config := Config{}
	if az.Config.CloudConfigType == "" || az.Config.CloudConfigType == cloudConfigTypeMerge {
		// Merge cloud config, set default value to existing config.
		config = az.Config
	}

	err = yaml.Unmarshal(cloudConfigData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Azure cloud-config: %w", err)
	}

	return &config, nil
}
