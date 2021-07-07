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
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestParseArchitecture(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "Intel Xeon Platinum 8259 (Cascade Lake)",
			expect: "amd64",
		},
		{
			input:  "AWS Graviton2 Processor",
			expect: "arm64",
		},
		{
			input:  "anything default",
			expect: "amd64",
		},
	}

	for _, test := range tests {
		got := parseArchitecture(test.input)
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

func TestUnmarshalProductsResponse(t *testing.T) {
	body := `
{
  "products": {
	"VVD8BG8WWFD3DAZN" : {
      "sku" : "VVD8BG8WWFD3DAZN",
      "productFamily" : "Compute Instance",
      "attributes" : {
        "servicecode" : "AmazonEC2",
        "location" : "US East (N. Virginia)",
        "locationType" : "AWS Region",
        "instanceType" : "r5b.4xlarge",
        "currentGeneration" : "Yes",
        "instanceFamily" : "Memory optimized",
        "vcpu" : "16",
        "physicalProcessor" : "Intel Xeon Platinum 8259 (Cascade Lake)",
        "clockSpeed" : "3.1 GHz",
        "memory" : "128 GiB",
        "storage" : "EBS only",
        "networkPerformance" : "Up to 10 Gigabit",
        "processorArchitecture" : "64-bit",
        "tenancy" : "Shared",
        "operatingSystem" : "Linux",
        "licenseModel" : "No License required",
        "usagetype" : "UnusedBox:r5b.4xlarge",
        "operation" : "RunInstances:0004",
        "availabilityzone" : "NA",
        "capacitystatus" : "UnusedCapacityReservation",
        "classicnetworkingsupport" : "false",
        "dedicatedEbsThroughput" : "10 Gbps",
        "ecu" : "NA",
        "enhancedNetworkingSupported" : "Yes",
        "instancesku" : "G4NFAXD9TGJM3RY8",
        "intelAvxAvailable" : "Yes",
        "intelAvx2Available" : "No",
        "intelTurboAvailable" : "No",
        "marketoption" : "OnDemand",
        "normalizationSizeFactor" : "32",
        "preInstalledSw" : "SQL Std",
        "servicename" : "Amazon Elastic Compute Cloud",
        "vpcnetworkingsupport" : "true"
      }
    },
    "C36QEQQQJ8ZR7N32" : {
      "sku" : "C36QEQQQJ8ZR7N32",
      "productFamily" : "Compute Instance",
      "attributes" : {
        "servicecode" : "AmazonEC2",
        "location" : "US East (N. Virginia)",
        "locationType" : "AWS Region",
        "instanceType" : "d3en.8xlarge",
        "currentGeneration" : "Yes",
        "instanceFamily" : "Storage optimized",
        "vcpu" : "32",
        "physicalProcessor" : "Intel Xeon Platinum 8259 (Cascade Lake)",
        "clockSpeed" : "3.1 GHz",
        "memory" : "128 GiB",
        "storage" : "16 x 14000 HDD",
        "networkPerformance" : "50 Gigabit",
        "processorArchitecture" : "64-bit",
        "tenancy" : "Dedicated",
        "operatingSystem" : "SUSE",
        "licenseModel" : "No License required",
        "usagetype" : "DedicatedRes:d3en.8xlarge",
        "operation" : "RunInstances:000g",
        "availabilityzone" : "NA",
        "capacitystatus" : "AllocatedCapacityReservation",
        "classicnetworkingsupport" : "false",
        "dedicatedEbsThroughput" : "5000 Mbps",
        "ecu" : "NA",
        "enhancedNetworkingSupported" : "Yes",
        "instancesku" : "2XW3BCEZ83WMGFJY",
        "intelAvxAvailable" : "Yes",
        "intelAvx2Available" : "Yes",
        "intelTurboAvailable" : "Yes",
        "marketoption" : "OnDemand",
        "normalizationSizeFactor" : "64",
        "preInstalledSw" : "NA",
        "processorFeatures" : "AVX; AVX2; Intel AVX; Intel AVX2; Intel AVX512; Intel Turbo",
        "servicename" : "Amazon Elastic Compute Cloud",
        "vpcnetworkingsupport" : "true"
      }
    }
  }
}
`
	r := strings.NewReader(body)
	resp, err := unmarshalProductsResponse(r)
	assert.Nil(t, err)
	assert.Len(t, resp.Products, 2)
	assert.NotNil(t, resp.Products["VVD8BG8WWFD3DAZN"])
	assert.NotNil(t, resp.Products["C36QEQQQJ8ZR7N32"])
	assert.Equal(t, resp.Products["VVD8BG8WWFD3DAZN"].Attributes.InstanceType, "r5b.4xlarge")
	assert.Equal(t, resp.Products["C36QEQQQJ8ZR7N32"].Attributes.InstanceType, "d3en.8xlarge")

	invalidJsonTests := map[string]string{
		"[":                     "[",
		"]":                     "]",
		"}":                     "}",
		"{":                     "{",
		"Plain text":            "invalid",
		"List":                  "[]",
		"Invalid products ([])": `{"products":[]}`,
		"Invalid product ([])":  `{"products":{"zz":[]}}`,
	}
	for name, body := range invalidJsonTests {
		t.Run(name, func(t *testing.T) {
			r := strings.NewReader(body)
			resp, err := unmarshalProductsResponse(r)
			assert.NotNil(t, err)
			assert.Nil(t, resp)
		})
	}
}
