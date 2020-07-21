/*
Copyright 2020 The Kubernetes Authors.

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

package pagination

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
)

// PageResult stores the HTTP response that returned the current page of results.
type PageResult struct {
	huaweicloudsdk.Result
	url.URL
}

// PageResultFrom parses an HTTP response as JSON and returns a PageResult containing the
// results, interpreting it as JSON if the content type indicates.
func PageResultFrom(resp *http.Response) (PageResult, error) {
	var parsedBody interface{}

	defer resp.Body.Close()
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PageResult{}, err
	}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		err = json.Unmarshal(rawBody, &parsedBody)
		if err != nil {
			return PageResult{}, err
		}
	} else {
		parsedBody = rawBody
	}

	return PageResultFromParsed(resp, parsedBody), err
}

// PageResultFromParsed constructs a PageResult from an HTTP response that has already had its
// body parsed as JSON (and closed).
func PageResultFromParsed(resp *http.Response, body interface{}) PageResult {
	return PageResult{
		Result: huaweicloudsdk.Result{
			Body:   body,
			Header: resp.Header,
		},
		URL: *resp.Request.URL,
	}
}

// Request performs an HTTP request and extracts the http.Response from the result.
func Request(client *huaweicloudsdk.ServiceClient, headers map[string]string, url string) (*http.Response, error) {
	return client.Get(url, nil, &huaweicloudsdk.RequestOpts{
		MoreHeaders: headers,
		OkCodes:     []int{200, 204, 300},
	})
}
