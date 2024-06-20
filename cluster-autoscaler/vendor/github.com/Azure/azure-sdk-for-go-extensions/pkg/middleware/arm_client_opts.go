/*
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

package middleware

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

func DefaultArmOpts(userAgent string, logCollector ArmRequestMetricCollector, customPerCallPolicies ...policy.Policy) *arm.ClientOptions {
	opts := &arm.ClientOptions{}
	opts.Telemetry = DefaultTelemetryOpts(userAgent)
	opts.Retry = DefaultRetryOpts()
	opts.Transport = DefaultHTTPClient()
	// we add the logging policy to the PerRetryPolicies so we can track
	// any retries that happened
	opts.PerRetryPolicies = []policy.Policy{
		runtime.NewRequestIDPolicy(),
		&ArmRequestMetricPolicy{Collector: logCollector},
	}
	opts.PerCallPolicies = []policy.Policy{}
	if customPerCallPolicies != nil {
		opts.PerCallPolicies = append(opts.PerCallPolicies, customPerCallPolicies...)
	}
	return opts
}

func DefaultRetryOpts() policy.RetryOptions {
	return policy.RetryOptions{
		MaxRetries: 5,
		// Note the default retry behavior is exponential backoff
		RetryDelay: time.Second * 5,
		// StatusCodes specifies the HTTP status codes that indicate the operation should be retried.
		// A nil slice will use the following values.
		//   http.StatusRequestTimeout      408
		//   http.StatusTooManyRequests     429
		//   http.StatusInternalServerError 500
		//   http.StatusBadGateway          502
		//   http.StatusServiceUnavailable  503
		//   http.StatusGatewayTimeout      504
		// Specifying values will replace the default values.
		// Specifying an empty slice will disable retries for HTTP status codes.
		// StatusCodes: nil,
	}
}

func DefaultTelemetryOpts(userAgent string) policy.TelemetryOptions {
	return policy.TelemetryOptions{
		ApplicationID: userAgent,
	}
}
