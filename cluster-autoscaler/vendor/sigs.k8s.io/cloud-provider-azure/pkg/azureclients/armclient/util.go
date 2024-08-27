/*
Copyright 2022 The Kubernetes Authors.

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

package armclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

func NewRateLimitSendDecorater(ratelimiter flowcontrol.RateLimiter, mc *metrics.MetricContext) autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if !ratelimiter.TryAccept() {
				mc.RateLimitedCount()
				return nil, fmt.Errorf("rate limit reached")
			}
			return s.Do(r)
		})
	}
}

func NewThrottledSendDecorater(mc *metrics.MetricContext) autorest.SendDecorator {
	var retryTimer time.Time
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if retryTimer.After(time.Now()) {
				mc.ThrottledCount()
				return nil, fmt.Errorf("request is throttled")
			}
			resp, err := s.Do(r)
			rerr := retry.GetError(resp, err)
			if rerr.IsThrottled() {
				// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
				retryTimer = rerr.RetryAfter
			}
			return resp, err
		})
	}
}

func NewErrorCounterSendDecorator(mc *metrics.MetricContext) autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			resp, err := s.Do(r)
			rerr := retry.GetError(resp, err)
			mc.Observe(rerr)
			return resp, err
		})
	}
}

func DoDumpRequest(v klog.Level) autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {

		return autorest.SenderFunc(func(request *http.Request) (*http.Response, error) {
			if request != nil {
				requestDump, err := httputil.DumpRequest(request, true)
				if err != nil {
					klog.Errorf("Failed to dump request: %v", err)
				} else {
					klog.V(v).Infof("Dumping request: %s", string(requestDump))
				}
			}
			return s.Do(request)
		})
	}
}

func WithMetricsSendDecoratorWrapper(prefix, request, resourceGroup, subscriptionID, source string, factory func(mc *metrics.MetricContext) []autorest.SendDecorator) autorest.SendDecorator {
	mc := metrics.NewMetricContext(prefix, request, resourceGroup, subscriptionID, source)
	if factory != nil {
		return func(s autorest.Sender) autorest.Sender {
			return autorest.DecorateSender(s, factory(mc)...)
		}
	}
	return nil
}

// DoHackRegionalRetryForGET checks if GET request returns empty response and retries regional server or returns error.
func DoHackRegionalRetryForGET(c *Client) autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(request *http.Request) (*http.Response, error) {
			response, rerr := s.Do(request)
			if response == nil {
				klog.V(2).Infof("response is empty")
				return response, rerr
			}

			bodyBytes, _ := io.ReadAll(response.Body)
			defer func() {
				response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}()

			bodyString := string(bodyBytes)
			trimmed := strings.TrimSpace(bodyString)
			klog.V(6).Infof("%s %s got response with ContentLength %d, StatusCode %d and responseBody length %d", request.Method, request.URL.Path, response.ContentLength, response.StatusCode, len(trimmed))

			// Hack: retry the regional ARM endpoint in case of ARM traffic split and arm resource group replication is too slow
			// Empty content and 2xx http status code are returned in this case.
			// Issue: https://github.com/kubernetes-sigs/cloud-provider-azure/issues/1296
			// Such situation also needs retrying that ContentLength is -1, StatusCode is 200 and an empty body is returned.
			emptyResp := (response.ContentLength == 0 || trimmed == "" || trimmed == "{}") && response.StatusCode >= 200 && response.StatusCode < 300
			if !emptyResp {
				if rerr == nil || response.StatusCode == http.StatusNotFound || c.regionalEndpoint == "" {
					return response, rerr
				}

				var body map[string]interface{}
				if e := json.Unmarshal(bodyBytes, &body); e != nil {
					klog.Errorf("Send.sendRequest: error in parsing response body string %q: %s, Skip retrying regional host", bodyBytes, e.Error())
					return response, rerr
				}

				err, ok := body["error"].(map[string]interface{})
				if !ok || err["code"] == nil || !strings.EqualFold(err["code"].(string), "ResourceGroupNotFound") {
					klog.V(5).Infof("Send.sendRequest: response body does not contain ResourceGroupNotFound error code. Skip retrying regional host")
					return response, rerr
				}
			}

			// Do regional request
			currentHost := request.URL.Host
			if request.Host != "" {
				currentHost = request.Host
			}

			if strings.HasPrefix(strings.ToLower(currentHost), c.regionalEndpoint) {
				klog.V(5).Infof("Send.sendRequest: current host %s is regional host. Skip retrying regional host.", html.EscapeString(currentHost))
				return response, rerr
			}

			request.Host = c.regionalEndpoint
			request.URL.Host = c.regionalEndpoint
			klog.V(6).Infof("Send.sendRegionalRequest on ResourceGroupNotFound error. Retrying regional host: %s", html.EscapeString(request.Host))

			regionalResponse, regionalError := s.Do(request)

			// only use the result if the regional request actually goes through and returns 2xx status code, for two reasons:
			// 1. the retry on regional ARM host approach is a hack.
			// 2. the concatenated regional uri could be wrong as the rule is not officially declared by ARM.
			if regionalResponse == nil || regionalResponse.StatusCode > 299 {
				regionalErrStr := ""
				if regionalError != nil {
					regionalErrStr = regionalError.Error()
				}

				klog.V(6).Infof("Send.sendRegionalRequest failed to get response from regional host, error: %q. Ignoring the result.", regionalErrStr)
				return response, rerr
			}

			// Do the same check on regional response just like the global one
			bodyBytes, _ = io.ReadAll(regionalResponse.Body)
			defer func() {
				regionalResponse.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}()
			bodyString = string(bodyBytes)
			trimmed = strings.TrimSpace(bodyString)
			emptyResp = (regionalResponse.ContentLength == 0 || trimmed == "" || trimmed == "{}") && regionalResponse.StatusCode >= 200 && regionalResponse.StatusCode < 300
			if emptyResp {
				contentLengthErrStr := fmt.Sprintf("empty response with trimmed body %q, ContentLength %d and StatusCode %d", trimmed, regionalResponse.ContentLength, regionalResponse.StatusCode)
				klog.Errorf(contentLengthErrStr)
				return response, fmt.Errorf(contentLengthErrStr)
			}

			return regionalResponse, regionalError
		})
	}
}
