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
	corev1 "k8s.io/api/core/v1"
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
// This acts as a compatibility layer with the legacy resource LimitsVal system.
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

// CombinedQuotasProvider wraps other Providers and combines their quotas.
type CombinedQuotasProvider struct {
	providers []Provider
}

// NewCombinedQuotasProvider returns a new CombinedQuotasProvider.
func NewCombinedQuotasProvider(providers []Provider) *CombinedQuotasProvider {
	return &CombinedQuotasProvider{
		providers: providers,
	}
}

// Quotas returns a union of quotas from all wrapped providers.
func (p *CombinedQuotasProvider) Quotas() ([]Quota, error) {
	var allQuotas []Quota
	for _, provider := range p.providers {
		quotas, err := provider.Quotas()
		if err != nil {
			return nil, err
		}
		allQuotas = append(allQuotas, quotas...)
	}
	return allQuotas, nil
}

// minQuotaAdapter wraps cloudprovider.ResourceLimiter and returns GetMin values.
type minQuotaAdapter struct {
	limiter *cloudprovider.ResourceLimiter
}

func (a *minQuotaAdapter) ID() string                       { return a.limiter.ID() }
func (a *minQuotaAdapter) AppliesTo(node *corev1.Node) bool { return a.limiter.AppliesTo(node) }
func (a *minQuotaAdapter) Limits() map[string]int64 {
	limits := make(map[string]int64)
	for _, r := range a.limiter.GetResources() {
		limits[r] = a.limiter.GetMin(r)
	}
	return limits
}

// maxQuotaAdapter wraps cloudprovider.ResourceLimiter and returns GetMax values.
type maxQuotaAdapter struct {
	limiter *cloudprovider.ResourceLimiter
}

func (a *maxQuotaAdapter) ID() string                       { return a.limiter.ID() }
func (a *maxQuotaAdapter) AppliesTo(node *corev1.Node) bool { return a.limiter.AppliesTo(node) }
func (a *maxQuotaAdapter) Limits() map[string]int64 {
	limits := make(map[string]int64)
	for _, r := range a.limiter.GetResources() {
		limits[r] = a.limiter.GetMax(r)
	}
	return limits
}

// CloudMinProvider provides minimum quotas from cloud provider.
type CloudMinProvider struct {
	cloudProvider cloudprovider.CloudProvider
}

// Quotas returns the minimum quotas from the cloud provider.
func (p *CloudMinProvider) Quotas() ([]Quota, error) {
	rl, err := p.cloudProvider.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	return []Quota{&minQuotaAdapter{limiter: rl}}, nil
}

// NewCloudMinProvider returns a new CloudMinProvider.
func NewCloudMinProvider(cloudProvider cloudprovider.CloudProvider) *CloudMinProvider {
	return &CloudMinProvider{cloudProvider: cloudProvider}
}

// CloudMaxProvider provides maximum quotas from cloud provider.
type CloudMaxProvider struct {
	cloudProvider cloudprovider.CloudProvider
}

// Quotas returns the maximum quotas from the cloud provider.
func (p *CloudMaxProvider) Quotas() ([]Quota, error) {
	rl, err := p.cloudProvider.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	return []Quota{&maxQuotaAdapter{limiter: rl}}, nil
}

// NewCloudMaxProvider returns a new CloudMaxProvider.
func NewCloudMaxProvider(cloudProvider cloudprovider.CloudProvider) *CloudMaxProvider {
	return &CloudMaxProvider{cloudProvider: cloudProvider}
}
