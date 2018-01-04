/*
Copyright 2017 The Kubernetes Authors.

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

package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var pricingBody []byte

func init() {
	f, err := os.Open("pricing_eu-west-1.json")
	if err != nil {
		panic(err)
	}
	pricingBody, err = ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
}

func TestInstanceInfoService_DescribeInstanceInfo(t *testing.T) {
	usEastOneURL, err := url.Parse(fmt.Sprintf(awsPricingAPIURLTemplate, "us-east-1"))
	assert.NoError(t, err)

	usWestOneURL, err := url.Parse(fmt.Sprintf(awsPricingAPIURLTemplate, "us-west-1"))
	assert.NoError(t, err)

	mc := &mockClient{m: make(map[string]mockResponse)}
	mc.m[usEastOneURL.Path] = mockResponse{pricingBody, 200}
	mc.m[usWestOneURL.Path] = mockResponse{[]byte("some non-json stuff"), 200}

	type testCase struct {
		instanceType        string
		region              string
		expectError         bool
		expectOnDemandPrice float64
		expectCPU           int64
	}
	type cases []testCase

	tcs := cases{
		{ // good case: common case
			"m4.xlarge",
			"us-east-1",
			false,
			0.2,
			4,
		},
		{ // error case: unknown availability region
			"m4.xlarge",
			"eu-east-2",
			true,
			0,
			0,
		},
		{ // error case: unknown instance
			"unknown-instance",
			"us-east-1",
			true,
			0,
			0,
		},
		{ // error case: invalid server response
			"m4.xlarge",
			"us-west-1",
			true,
			0,
			0,
		},
	}

	service := NewEC2InstanceInfoService(mc)

	for n, tc := range tcs {
		info, err := service.DescribeInstanceInfo(tc.instanceType, tc.region)
		if tc.expectError {
			assert.Error(t, err, fmt.Sprintf("case %d", n))
		} else {
			assert.NoError(t, err, fmt.Sprintf("case %d", n))
			assert.Equal(t, tc.instanceType, info.InstanceType)
			assert.Equal(t, tc.expectCPU, info.VCPU)
			assert.Equal(t, tc.expectOnDemandPrice, info.OnDemandPrice)
		}

	}
}

type mockResponse struct {
	body       []byte
	statusCode int
}

type mockClient struct {
	m map[string]mockResponse
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	if mock, found := m.m[req.URL.Path]; found {
		return &http.Response{
			Status:        http.StatusText(mock.statusCode),
			StatusCode:    mock.statusCode,
			ContentLength: int64(len(mock.body)),
			Body:          ioutil.NopCloser(bytes.NewReader(mock.body)),
			Request:       req,
		}, nil
	}
	return &http.Response{
		Status:        http.StatusText(404),
		StatusCode:    404,
		Request:       req,
		Body:          ioutil.NopCloser(bytes.NewReader([]byte{})),
		ContentLength: 0,
	}, nil
}
