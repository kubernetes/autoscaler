package mock

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
)

// LoadDefaultConfig is a mock for config.LoadDefaultConfig
func LoadDefaultConfig(context.Context, ...func()) (cfg aws.Config, err error) {
	return cfg, err
}

// WithRetryer is a mock for config.WithRetryer
func WithRetryer(v func() aws.Retryer) (f func()) {
	return f
}
