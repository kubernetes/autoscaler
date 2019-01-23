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
	"fmt"
	"os"
	"testing"

	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/stretchr/testify/assert"
)

func loadMockData(t *testing.T) []aws.JSONValue {
	f, err := os.Open("pricing_ondemand_eu-west-1.json")
	if err != nil {
		t.Fatalf("Failed to open mock file: %v", err)
	}

	grütze := &pricing.GetProductsOutput{}

	err = jsonutil.UnmarshalJSON(grütze, f)
	if err != nil {
		t.Fatalf("Failed transform mock JSON into AWS JSONValue: %v", err)
	}

	return grütze.PriceList
}

func TestInstanceInfoService_DescribeInstanceInfo(t *testing.T) {
	tcs := []struct {
		name                  string
		instanceType          string
		region                string
		data                  []aws.JSONValue
		errorExpected         bool
		expectedError         string
		expectedOnDemandPrice float64
		expectedCPU           int64
	}{
		{
			name:                  "error case: unknown availability region",
			instanceType:          "m4.xlarge",
			region:                "unknown-region",
			data:                  []aws.JSONValue{},
			errorExpected:         true,
			expectedError:         "region full name not found for region: unknown-region",
			expectedOnDemandPrice: 0,
			expectedCPU:           0,
		},
		{
			name:                  "error case: invalid server response",
			instanceType:          "m4.xlarge",
			region:                "us-west-1",
			data:                  []aws.JSONValue{},
			errorExpected:         true,
			expectedError:         "failed to sync aws product and price information: no price information found for region us-west-1",
			expectedOnDemandPrice: 0,
			expectedCPU:           0,
		},
		{
			name:                  "error case: unknown instance",
			instanceType:          "unknown-instance",
			region:                "eu-west-1",
			data:                  loadMockData(t),
			errorExpected:         true,
			expectedError:         "instance info not available for instance type unknown-instance region eu-west-1",
			expectedOnDemandPrice: 0,
			expectedCPU:           0,
		},
		{
			name:                  "good case: common case",
			instanceType:          "m4.xlarge",
			region:                "eu-west-1",
			data:                  loadMockData(t),
			errorExpected:         false,
			expectedOnDemandPrice: 0.222,
			expectedCPU:           4,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			indexRegion, err := regionFullName(tc.region)
			if err != nil {
				assert.Equal(t, err.Error(), tc.expectedError)
				return
			}

			mc := &mockClient{m: make(map[string][]aws.JSONValue)}
			mc.m[indexRegion] = tc.data

			service := NewEC2InstanceInfoService(mc)
			info, err := service.DescribeInstanceInfo(tc.instanceType, tc.region)

			if tc.errorExpected {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.instanceType, info.InstanceType)
				assert.Equal(t, tc.expectedCPU, info.VCPU)
				assert.Equal(t, tc.expectedOnDemandPrice, info.OnDemandPrice)
			}
		})
	}
}

type mockClient struct {
	m map[string][]aws.JSONValue
}

func (m *mockClient) GetProducts(input *pricing.GetProductsInput) (*pricing.GetProductsOutput, error) {
	region := getRegionFromFilters(input.Filters)

	if mock, found := m.m[region]; found {
		return &pricing.GetProductsOutput{
			PriceList: mock,
		}, nil
	}

	return nil, fmt.Errorf("no price information found for region %s", region)
}

func getRegionFromFilters(filters []*pricing.Filter) string {
	for _, filter := range filters {
		if reflect.DeepEqual(filter.Field, aws.String("location")) {
			return aws.StringValue(filter.Value)
		}
	}

	return "no-region"
}
