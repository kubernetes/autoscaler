/*
Copyright 2021 The Kubernetes Authors.

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

package common

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/profile"
)

const (
	tplNetworkFailureRetry = "[WARN] temporary network failure, retrying (%d/%d) in %f seconds: %s"
)

func (c *Client) sendWithNetworkFailureRetry(req *http.Request, retryable bool) (resp *http.Response, err error) {
	// make sure maxRetries is more than or equal 0
	var maxRetries int
	if retryable {
		maxRetries = maxInt(c.profile.NetworkFailureMaxRetries, 0)
	}
	durationFunc := safeDurationFunc(c.profile.NetworkFailureRetryDuration)

	for idx := 0; idx <= maxRetries; idx++ {
		resp, err = c.sendHttp(req)

		// retry when error occurred and retryable and not the last retry
		// should not sleep on last retry even if it's retryable
		if err != nil && retryable && idx < maxRetries {
			if err, ok := err.(net.Error); ok && (err.Timeout() || err.Temporary()) {
				duration := durationFunc(idx)
				if c.debug {
					c.logger.Printf(tplNetworkFailureRetry, idx, maxRetries, duration.Seconds(), err.Error())
				}

				time.Sleep(duration)
				continue
			}
		}

		if err != nil {
			msg := fmt.Sprintf("Fail to get response because %s", err)
			err = errors.NewTencentCloudSDKError("ClientError.NetworkError", msg, "")
		}

		return resp, err
	}

	return resp, err
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func safeDurationFunc(durationFunc profile.DurationFunc) profile.DurationFunc {
	if durationFunc != nil {
		return durationFunc
	}
	return profile.ExponentialBackoff
}

// isRetryable means if request is retryable or not,
// depends on if request has a `ClientToken` field or not,
// request with `ClientToken` means it's idempotent and retryable,
// unretryable request SHOULDN'T retry for temporary network failure
func isRetryable(obj interface{}) bool {
	// obj Must be struct ptr
	getType := reflect.TypeOf(obj)
	if getType.Kind() != reflect.Ptr || getType.Elem().Kind() != reflect.Struct {
		return false
	}

	// obj Must exist named field
	_, ok := getType.Elem().FieldByName(fieldClientToken)
	return ok
}
