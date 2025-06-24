package mock

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
)

// Options is a mock client Options
type Options struct {
	Retryer aws.Retryer
}

// Client is a mock service client
type Client struct{}

// GetObjectInput is mock input
type GetObjectInput struct {
	Bucket *string
	Key    *string
}

// GetObjectOutput is mock output
type GetObjectOutput struct{}

// NewFromConfig is a mock client constructor
func NewFromConfig(aws.Config, ...func(options *Options)) Client {
	return Client{}
}

// GetObject is a mock GetObject API
func (Client) GetObject(context.Context, *GetObjectInput, ...func(*Options)) (o *GetObjectOutput, err error) {
	return o, err
}
