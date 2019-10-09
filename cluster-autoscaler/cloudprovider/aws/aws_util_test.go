/*
Copyright 2019 The Kubernetes Authors.

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

package aws

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestGetStaticEC2InstanceTypes(t *testing.T) {
	result, _ := GetStaticEC2InstanceTypes()
	assert.True(t, len(result) != 0)
}

func TestParseMemory(t *testing.T) {
	expectedResultInMiB := int64(3.75 * 1024)
	tests := []struct {
		input  string
		expect int64
	}{
		{
			input:  "3.75 GiB",
			expect: expectedResultInMiB,
		},
		{
			input:  "3.75 Gib",
			expect: expectedResultInMiB,
		},
		{
			input:  "3.75GiB",
			expect: expectedResultInMiB,
		},
		{
			input:  "3.75",
			expect: expectedResultInMiB,
		},
	}

	for _, test := range tests {
		got := parseMemory(test.input)
		assert.Equal(t, test.expect, got)
	}
}

func TestParseCPU(t *testing.T) {
	tests := []struct {
		input  string
		expect int64
	}{
		{
			input:  strconv.FormatInt(8, 10),
			expect: int64(8),
		},
	}

	for _, test := range tests {
		got := parseCPU(test.input)
		assert.Equal(t, test.expect, got)
	}
}

func TestGetCurrentAwsRegion(t *testing.T) {
	region := "us-west-2"
	if oldRegion, found := os.LookupEnv("AWS_REGION"); found {
		os.Unsetenv("AWS_REGION")
		defer os.Setenv("AWS_REGION", oldRegion)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("{\"region\" : \"" + region + "\"}"))
	}))
	// Close the server when test finishes
	defer server.Close()

	ec2MetaDataServiceUrl = server.URL
	result, err := GetCurrentAwsRegion()

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, region, result)
}

func TestGetCurrentAwsRegionWithRegionEnv(t *testing.T) {
	region := "us-west-2"
	if oldRegion, found := os.LookupEnv("AWS_REGION"); found {
		defer os.Setenv("AWS_REGION", oldRegion)
	} else {
		defer os.Unsetenv("AWS_REGION")
	}
	os.Setenv("AWS_REGION", region)

	result, err := GetCurrentAwsRegion()
	assert.Nil(t, err)
	assert.Equal(t, region, result)
}
