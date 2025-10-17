package resourcequotas

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type Provider interface {
	Quotas() ([]Quota, error)
}
type CloudQuotasProvider struct {
	cloudProvider cloudprovider.CloudProvider
}

func (p *CloudQuotasProvider) Quotas() ([]Quota, error) {
	rl, err := p.cloudProvider.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	return []Quota{rl}, nil
}

func NewCloudQuotasProvider(cloudProvider cloudprovider.CloudProvider) *CloudQuotasProvider {
	return &CloudQuotasProvider{
		cloudProvider: cloudProvider,
	}
}
