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

package utils

import (
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/policy/retryrepectthrottled"
)

var TracingProvider tracing.Provider

const (
	DefaultMaxRetries    = 3
	DefaultMaxRetryDelay = 60 * time.Second
	DefaultRetryDelay    = 5 * time.Second
)

func GetDefaultOption() *arm.ClientOptions {
	return &arm.ClientOptions{
		ClientOptions: GetDefaultAzCoreClientOption(),
	}
}

func GetDefaultAzCoreClientOption() policy.ClientOptions {
	return policy.ClientOptions{
		Retry: policy.RetryOptions{
			RetryDelay:    DefaultRetryDelay,
			MaxRetryDelay: DefaultMaxRetryDelay,
			MaxRetries:    DefaultMaxRetries,
			StatusCodes:   retryrepectthrottled.GetRetriableStatusCode(),
		},
		PerRetryPolicies: []policy.Policy{
			retryrepectthrottled.NewThrottlingPolicy(),
		},
		Transport: &http.Client{
			Transport: DefaultTransport,
		},
		TracingProvider:                 TracingProvider,
		Cloud:                           cloud.AzurePublic,
		InsecureAllowCredentialWithHTTP: true,
	}
}
