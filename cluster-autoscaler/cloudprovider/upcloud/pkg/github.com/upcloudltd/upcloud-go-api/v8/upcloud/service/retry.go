package service

import (
	"context"
	"errors"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/client"
)

const maxRetriesOnError = 3

type retryConfig struct {
	interval time.Duration
	// Inverse the should retry logic. By default, operation is retried until operation returns a value. If inverse is set to true, operation is retried while operation returns a value. This should be used, for example, for waiting until resource is deleted.
	inverse bool
}

func fillDefaults(c *retryConfig) *retryConfig {
	if c == nil {
		c = &retryConfig{}
	}

	if c.interval.Milliseconds() == 0 {
		c.interval = time.Second * 5
	}

	return c
}

func shouldRetryOnError(err error, retryOnErrorCount int) bool {
	if retryOnErrorCount >= maxRetriesOnError {
		return false
	}

	var ucErr *upcloud.Problem
	if errors.As(err, &ucErr) && ucErr.Status >= 500 {
		return true
	}

	var clientErr *client.Error
	if errors.As(err, &clientErr) && clientErr.ErrorCode >= 500 {
		return true
	}

	return false
}

func retry[T any](ctx context.Context, operation func(int, context.Context) (*T, error), config *retryConfig) (*T, error) {
	config = fillDefaults(config)

	ticker := time.NewTicker(config.interval)
	defer ticker.Stop()

	retryOnErrorCount := 0
	for i := 0; ; i++ {
		select {
		case <-ticker.C:
			value, err := operation(i, ctx)

			if err != nil {
				if shouldRetryOnError(err, retryOnErrorCount) {
					retryOnErrorCount++
					continue
				}

				return value, err
			} else {
				retryOnErrorCount = 0
			}

			if !config.inverse && value != nil {
				return value, nil
			}

			if config.inverse && value == nil {
				return nil, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
