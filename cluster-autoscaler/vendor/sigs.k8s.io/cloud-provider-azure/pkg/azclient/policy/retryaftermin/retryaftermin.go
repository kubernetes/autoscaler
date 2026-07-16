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

package retryaftermin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"k8s.io/klog/v2"
)

// Policy is a policy that enforces a minimum retry-after value
type Policy struct {
	minRetryAfter time.Duration
}

// NewRetryAfterMinPolicy creates a new Policy with the specified minimum retry-after value
func NewRetryAfterMinPolicy(minRetryAfter time.Duration) policy.Policy {
	return &Policy{
		minRetryAfter: minRetryAfter,
	}
}

// GetMinRetryAfter returns the minimum retry-after value
func (p *Policy) GetMinRetryAfter() time.Duration {
	return p.minRetryAfter
}

// Do implements the policy.Policy interface
func (p *Policy) Do(req *policy.Request) (*http.Response, error) {
	logger := klog.Background().WithName("RetryAfterMinPolicy.Do")
	resp, err := req.Next()
	// If the request failed or the status code is >= 300, return
	if err != nil || resp == nil || resp.StatusCode >= 300 {
		return resp, err
	}

	// Check if the response retry-after header is less than the minimum
	overrideRetryAfter := func(header http.Header, headerName string, retryAfter time.Duration) {
		if retryAfter < p.minRetryAfter {
			logger.V(5).Info("retry-after value is less than minimum, removing retry-after header", "retryAfter", retryAfter, "minimum", p.minRetryAfter)
			header.Del(headerName)
		}
	}

	// Process all "Retry-After" headers (case-insensitive)
	for headerName := range resp.Header {
		if strings.EqualFold(headerName, "Retry-After") {
			retryAfter := resp.Header.Get(headerName)
			// Try to parse as duration (e.g., "1s", "30s", "1m")
			retryDuration, err := time.ParseDuration(retryAfter)
			if err == nil {
				// If the retry-after value is less than the minimum, remove it
				overrideRetryAfter(resp.Header, headerName, retryDuration)
			} else {
				// Try to parse as seconds (integer)
				seconds, err := strconv.Atoi(retryAfter)
				if err == nil {
					retryDuration := time.Duration(seconds) * time.Second
					// If the retry-after value is less than the minimum, remove it
					overrideRetryAfter(resp.Header, headerName, retryDuration)
				} else {
					logger.V(5).Info("Not modifying header with unrecognized format", "headerName", headerName, "unrecognized format", retryAfter)
				}
			}
		}
	}

	return resp, err
}
