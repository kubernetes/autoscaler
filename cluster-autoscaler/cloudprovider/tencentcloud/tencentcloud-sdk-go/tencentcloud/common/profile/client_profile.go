/*
Copyright 2016 The Kubernetes Authors.

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

package profile

import (
	"math"
	"time"
)

type DurationFunc func(index int) time.Duration

func ConstantDurationFunc(duration time.Duration) DurationFunc {
	return func(int) time.Duration {
		return duration
	}
}

func ExponentialBackoff(index int) time.Duration {
	seconds := math.Pow(2, float64(index))
	return time.Duration(seconds) * time.Second
}

type ClientProfile struct {
	HttpProfile *HttpProfile
	// Valid choices: HmacSHA1, HmacSHA256, TC3-HMAC-SHA256.
	// Default value is TC3-HMAC-SHA256.
	SignMethod      string
	UnsignedPayload bool
	// Valid choices: zh-CN, en-US.
	// Default value is zh-CN.
	Language string
	Debug    bool
	// define Whether to enable Regional auto switch
	DisableRegionBreaker bool

	// Deprecated. Use BackupEndpoint instead.
	BackupEndPoint string
	BackupEndpoint string

	// define how to retry request
	NetworkFailureMaxRetries       int
	NetworkFailureRetryDuration    DurationFunc
	RateLimitExceededMaxRetries    int
	RateLimitExceededRetryDuration DurationFunc
}

func NewClientProfile() *ClientProfile {
	return &ClientProfile{
		HttpProfile:     NewHttpProfile(),
		SignMethod:      "TC3-HMAC-SHA256",
		UnsignedPayload: false,
		Language:        "zh-CN",
		Debug:           false,
		// now is true, will become to false in future
		DisableRegionBreaker: true,
		BackupEndPoint:       "",
		BackupEndpoint:       "",
	}
}
