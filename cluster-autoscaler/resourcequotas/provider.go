/*
Copyright 2025 The Kubernetes Authors.

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

package resourcequotas

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// Provider provides Quotas. Each Provider implementation acts as a different
// source of Quotas.
type Provider interface {
	Quotas() ([]Quota, error)
}

// CloudQuotasProvider is an adapter for cloudprovider.ResourceLimiter.
type CloudQuotasProvider struct {
	cloudProvider cloudprovider.CloudProvider
}

// Quotas returns the cloud provider's ResourceLimiter, which implements Quota interface.
//
// This acts as a compatibility layer with the legacy resource limits system.
func (p *CloudQuotasProvider) Quotas() ([]Quota, error) {
	rl, err := p.cloudProvider.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	return []Quota{rl}, nil
}

// NewCloudQuotasProvider returns a new CloudQuotasProvider.
func NewCloudQuotasProvider(cloudProvider cloudprovider.CloudProvider) *CloudQuotasProvider {
	return &CloudQuotasProvider{
		cloudProvider: cloudProvider,
	}
}
