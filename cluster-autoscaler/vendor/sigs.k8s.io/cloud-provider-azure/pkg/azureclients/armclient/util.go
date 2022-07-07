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
	"io/ioutil"
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

// DoExponentialBackoffRetry returns an autorest.SendDecorator which performs retry with customizable backoff policy.
func DoHackRegionalRetryDecorator(c *Client) autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(request *http.Request) (*http.Response, error) {
			response, rerr := s.Do(request)
			if response == nil {
				klog.V(2).Infof("response is empty")
				return response, rerr
			}
			if rerr == nil || response.StatusCode == http.StatusNotFound || c.regionalEndpoint == "" {
				return response, rerr
			}
			// Hack: retry the regional ARM endpoint in case of ARM traffic split and arm resource group replication is too slow
			bodyBytes, _ := ioutil.ReadAll(response.Body)
			defer func() {
				response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			}()

			bodyString := string(bodyBytes)
			var body map[string]interface{}
			if e := json.Unmarshal(bodyBytes, &body); e != nil {
				klog.Errorf("Send.sendRequest: error in parsing response body string: %s, Skip retrying regional host", e.Error())
				return response, rerr
			}
			klog.V(5).Infof("Send.sendRequest original response: %s", bodyString)

			if err, ok := body["error"].(map[string]interface{}); !ok ||
				err["code"] == nil ||
				!strings.EqualFold(err["code"].(string), "ResourceGroupNotFound") {
				klog.V(5).Infof("Send.sendRequest: response body does not contain ResourceGroupNotFound error code. Skip retrying regional host")
				return response, rerr
			}

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
			klog.V(5).Infof("Send.sendRegionalRequest on ResourceGroupNotFound error. Retrying regional host: %s", html.EscapeString(request.Host))

			regionalResponse, regionalError := s.Do(request)
			// only use the result if the regional request actually goes through and returns 2xx status code, for two reasons:
			// 1. the retry on regional ARM host approach is a hack.
			// 2. the concatenated regional uri could be wrong as the rule is not officially declared by ARM.
			if regionalResponse == nil || regionalResponse.StatusCode > 299 {
				regionalErrStr := ""
				if regionalError != nil {
					regionalErrStr = regionalError.Error()
				}

				klog.V(5).Infof("Send.sendRegionalRequest failed to get response from regional host, error: '%s'. Ignoring the result.", regionalErrStr)
				return response, rerr
			}
			return regionalResponse, regionalError
		})
	}
}
