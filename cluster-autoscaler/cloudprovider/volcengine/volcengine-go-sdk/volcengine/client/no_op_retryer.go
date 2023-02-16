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

package client

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
)

// NoOpRetryer provides a retryer that performs no retries.
// It should be used when we do not want retries to be performed.
type NoOpRetryer struct{}

// MaxRetries returns the number of maximum returns the service will use to make
// an individual API; For NoOpRetryer the MaxRetries will always be zero.
func (d NoOpRetryer) MaxRetries() int {
	return 0
}

// ShouldRetry will always return false for NoOpRetryer, as it should never retry.
func (d NoOpRetryer) ShouldRetry(_ *request.Request) bool {
	return false
}

// RetryRules returns the delay duration before retrying this request again;
// since NoOpRetryer does not retry, RetryRules always returns 0.
func (d NoOpRetryer) RetryRules(_ *request.Request) time.Duration {
	return 0
}
