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

package retryrepectthrottled

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

const HeaderRetryAfter = "Retry-After"

func NewThrottlingPolicy() policy.Policy {
	return &ThrottlingPolicy{
		RetryAfterReader: time.Now(),
		RetryAfterWriter: time.Now(),
	}
}

func GetRetriableStatusCode() []int {
	return []int{
		http.StatusRequestTimeout,      // 408
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
	}
}

// ThrottlingPolicy implements the Azure SDK for Go's Policy interface.
// throttle counter is based on resources operation per subscription.
type ThrottlingPolicy struct {
	RetryAfterReader time.Time
	RetryAfterWriter time.Time
}

func (p *ThrottlingPolicy) Do(req *policy.Request) (*http.Response, error) {
	if req.Raw().Method == http.MethodGet || req.Raw().Method == http.MethodHead {
		return p.processThrottlePolicy(&p.RetryAfterReader, req)
	}
	return p.processThrottlePolicy(&p.RetryAfterWriter, req)
}

func (p *ThrottlingPolicy) processThrottlePolicy(timer *time.Time, req *policy.Request) (*http.Response, error) {
	if timer.After(time.Now()) {
		return nil, errors.New("ThrottlingPolicy: Too many requests")
	}
	resp, err := req.Next()
	if err != nil {
		return resp, err
	}
	if runtime.HasStatusCode(resp, http.StatusTooManyRequests) {
		// throttle policy will be triggered when the response status code is 429
		// in v1 client, throttle policy will be triggered when the retry-after header is set
		// according to https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/async-operations
		// the retry-after header will be set when the response status code is 202
		duration := resp.Header.Get(HeaderRetryAfter)
		if duration == "" {
			*timer = time.Now()
		}
		if retryAfter, _ := strconv.Atoi(duration); retryAfter > 0 {
			*timer = time.Now().Add(time.Duration(retryAfter) * time.Second)
		} else if t, err := time.Parse(time.RFC1123, duration); err == nil {
			*timer = t
		}

		return resp, errors.New("ThrottlingPolicy: Too many requests")
	}
	return resp, nil
}
