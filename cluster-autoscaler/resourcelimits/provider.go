package resourcelimits

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type Provider interface {
	AllLimiters() ([]Limiter, error)
}
type CloudLimitersProvider struct {
	cloudProvider cloudprovider.CloudProvider
}

func (p *CloudLimitersProvider) AllLimiters() ([]Limiter, error) {
	rl, err := p.cloudProvider.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	return []Limiter{rl}, nil
}

func NewCloudLimitersProvider(cloudProvider cloudprovider.CloudProvider) *CloudLimitersProvider {
	return &CloudLimitersProvider{
		cloudProvider: cloudProvider,
	}
}
