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

/*
Functions to handle Kamatera API calls
Copied from the Kamatera terraform provider:
https://github.com/Kamatera/terraform-provider-kamatera/blob/master/kamatera/request.go
*/

package kamatera

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

var kamateraHTTPClient = &http.Client{Timeout: 5 * time.Minute}

// ProviderConfig is the configuration for the Kamatera cloud provider
type ProviderConfig struct {
	ApiUrl      string
	ApiClientID string
	ApiSecret   string
}

func request(ctx context.Context, provider ProviderConfig, method string, path string, body interface{}, numRetries int, secondsBetweenRetries int) (interface{}, error) {
	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		payload = b
	}
	path = strings.TrimPrefix(path, "/")
	logLevel := klog.Level(2)
	if strings.HasPrefix(path, "service/queue") || (method == "GET" && path == "service/servers") {
		logLevel = klog.Level(4)
	}
	url := fmt.Sprintf("%s/%s", provider.ApiUrl, path)
	var result interface{}
	var err error
	for attempt := 0; attempt < numRetries; attempt++ {
		result = nil
		err = nil
		klog.V(logLevel).Infof("kamatera request: %s %s %s", method, url, string(payload))
		var r io.Reader
		if payload != nil {
			r = bytes.NewReader(payload) // NEW reader each try
		}
		if attempt > 0 {
			klog.V(logLevel).Infof("kamatera request retry %d", attempt)
			time.Sleep(time.Duration(secondsBetweenRetries<<attempt) * time.Second)
		}
		req, e := http.NewRequestWithContext(ctx, method, url, r)
		if e != nil {
			err = e
			continue
		}
		req.Header.Add("AuthClientId", provider.ApiClientID)
		req.Header.Add("AuthSecret", provider.ApiSecret)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		res, e := kamateraHTTPClient.Do(req)
		if e != nil {
			err = e
			continue
		}
		e = func() error {
			defer res.Body.Close()
			decErr := json.NewDecoder(res.Body).Decode(&result)
			if decErr != nil {
				if res.StatusCode != 200 {
					return fmt.Errorf("bad status code from Kamatera API: %d", res.StatusCode)
				}
				return fmt.Errorf("invalid response from Kamatera API: %+v", result)
			} else if res.StatusCode != 200 {
				return fmt.Errorf("error response from Kamatera API (%d): %+v", res.StatusCode, result)
			}
			return nil
		}()
		if e != nil {
			err = e
			continue
		}
		err = nil
		break
	}
	return result, err
}
