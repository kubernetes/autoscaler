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

package ratelimit

import (
	"errors"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/policy/ratelimit/flowcontrol"
)

// Config indicates the rate limit config options.
type Config struct {
	// Enable rate limiting
	CloudProviderRateLimit bool `json:"cloudProviderRateLimit,omitempty" yaml:"cloudProviderRateLimit,omitempty"`
	// Rate limit QPS (Read)
	CloudProviderRateLimitQPS float32 `json:"cloudProviderRateLimitQPS,omitempty" yaml:"cloudProviderRateLimitQPS,omitempty"`
	// Rate limit Bucket Size
	CloudProviderRateLimitBucket int `json:"cloudProviderRateLimitBucket,omitempty" yaml:"cloudProviderRateLimitBucket,omitempty"`
	// Rate limit QPS (Write)
	CloudProviderRateLimitQPSWrite float32 `json:"cloudProviderRateLimitQPSWrite,omitempty" yaml:"cloudProviderRateLimitQPSWrite,omitempty"`
	// Rate limit Bucket Size
	CloudProviderRateLimitBucketWrite int `json:"cloudProviderRateLimitBucketWrite,omitempty" yaml:"cloudProviderRateLimitBucketWrite,omitempty"`
}

func NewRateLimitPolicy(config *Config) policy.Policy {
	if config != nil && config.CloudProviderRateLimit {
		readLimiter := flowcontrol.NewTokenBucketRateLimiter(
			config.CloudProviderRateLimitQPS,
			config.CloudProviderRateLimitBucket)

		writeLimiter := flowcontrol.NewTokenBucketRateLimiter(
			config.CloudProviderRateLimitQPSWrite,
			config.CloudProviderRateLimitBucketWrite)
		return &Policy{
			rateLimiterReader: readLimiter,
			rateLimiterWriter: writeLimiter,
		}
	}
	return nil
}

type Policy struct {
	rateLimiterWriter flowcontrol.RateLimiter
	rateLimiterReader flowcontrol.RateLimiter
}

func (f Policy) Do(req *policy.Request) (*http.Response, error) {
	if req.Raw().Method == http.MethodGet || req.Raw().Method == http.MethodHead {
		if !f.rateLimiterReader.TryAccept() {
			return nil, errors.New("rate limit reached")
		}
	} else {
		if !f.rateLimiterWriter.TryAccept() {
			return nil, errors.New("rate limit reached")
		}
	}
	return req.Next()
}

// CloudProviderRateLimitConfig indicates the rate limit config for each clients.
type CloudProviderRateLimitConfig struct {
	// The default rate limit config options.
	Config

	// Rate limit config for each clients. Values would override default settings above.
	Entries map[string]*Config `json:",inline" yaml:",inline"`
}

func NewCloudProviderRateLimitConfig() *CloudProviderRateLimitConfig {
	return &CloudProviderRateLimitConfig{
		Config: Config{
			CloudProviderRateLimit: false,
		},
		Entries: make(map[string]*Config),
	}
}

// GetRateLimitConfig returns the rate limit config for the given client. if the client is not found, the default is returned.
func (config *CloudProviderRateLimitConfig) GetRateLimitConfig(clientName string) *Config {
	if entry, ok := config.Entries[clientName]; ok {
		return entry
	}
	return &config.Config
}
