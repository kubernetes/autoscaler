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

package verda

import (
	"context"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

func TestInstanceTypesService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get instance types with default currency", func(t *testing.T) {
		ctx := context.Background()
		instanceTypes, err := client.InstanceTypes.Get(ctx, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instanceTypes) == 0 {
			t.Error("expected at least one instance type")
		}

		// Verify first instance type has expected fields
		if len(instanceTypes) > 0 {
			it := instanceTypes[0]
			if it.ID == "" {
				t.Error("expected instance type to have an ID")
			}
			if it.InstanceType == "" {
				t.Error("expected instance type to have an InstanceType")
			}
			if it.Model == "" {
				t.Error("expected instance type to have a Model")
			}
			if it.Name == "" {
				t.Error("expected instance type to have a Name")
			}
			if it.Currency == "" {
				t.Error("expected instance type to have a Currency")
			}
			if it.Manufacturer == "" {
				t.Error("expected instance type to have a Manufacturer")
			}
		}
	})

	t.Run("get instance types with USD currency", func(t *testing.T) {
		ctx := context.Background()
		instanceTypes, err := client.InstanceTypes.Get(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instanceTypes) == 0 {
			t.Error("expected at least one instance type")
		}
	})

	t.Run("get instance types with EUR currency", func(t *testing.T) {
		ctx := context.Background()
		instanceTypes, err := client.InstanceTypes.Get(ctx, "eur")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instanceTypes) == 0 {
			t.Error("expected at least one instance type")
		}
	})

	t.Run("verify flexible float for prices", func(t *testing.T) {
		ctx := context.Background()
		instanceTypes, err := client.InstanceTypes.Get(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instanceTypes) > 0 {
			it := instanceTypes[0]
			// Verify FlexibleFloat conversion works
			pricePerHour := it.PricePerHour.Float64()
			if pricePerHour <= 0 {
				t.Error("expected price per hour to be positive")
			}

			spotPrice := it.SpotPrice.Float64()
			if spotPrice <= 0 {
				t.Error("expected spot price to be positive")
			}

			dynamicPrice := it.DynamicPrice.Float64()
			if dynamicPrice <= 0 {
				t.Error("expected dynamic price to be positive")
			}

			maxDynamicPrice := it.MaxDynamicPrice.Float64()
			if maxDynamicPrice <= 0 {
				t.Error("expected max dynamic price to be positive")
			}
		}
	})

	t.Run("verify instance type has hardware specs", func(t *testing.T) {
		ctx := context.Background()
		instanceTypes, err := client.InstanceTypes.Get(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instanceTypes) > 0 {
			it := instanceTypes[0]
			// Check CPU specs
			if it.CPU.NumberOfCores <= 0 {
				t.Error("expected CPU to have number of cores")
			}
			// Check GPU specs
			if it.GPU.NumberOfGPUs <= 0 {
				t.Error("expected GPU to have number of GPUs")
			}
			// Check Memory specs
			if it.Memory.SizeInGigabytes <= 0 {
				t.Error("expected Memory to have size in gigabytes")
			}
			// Check GPU Memory specs
			if it.GPUMemory.SizeInGigabytes <= 0 {
				t.Error("expected GPU Memory to have size in gigabytes")
			}
		}
	})

	t.Run("verify instance type has best_for recommendations", func(t *testing.T) {
		ctx := context.Background()
		instanceTypes, err := client.InstanceTypes.Get(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instanceTypes) > 0 {
			it := instanceTypes[0]
			if it.BestFor == nil {
				t.Error("expected instance type to have BestFor field (can be empty slice)")
			}
		}
	})
}

func TestInstanceTypesService_GetByInstanceType(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get H100 instance type", func(t *testing.T) {
		ctx := context.Background()
		instanceType, err := client.InstanceTypes.GetByInstanceType(ctx, "1H100.80S.22V", false, "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instanceType == nil {
			t.Fatal("expected instance type, got nil")
		}

		if instanceType.InstanceType != "1H100.80S.22V" {
			t.Errorf("expected instance type 1H100.80S.22V, got %s", instanceType.InstanceType)
		}

		if instanceType.Model != "H100" {
			t.Errorf("expected model H100, got %s", instanceType.Model)
		}
	})

	t.Run("get V100 instance type", func(t *testing.T) {
		ctx := context.Background()
		instanceType, err := client.InstanceTypes.GetByInstanceType(ctx, "1V100.6V", false, "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instanceType == nil {
			t.Fatal("expected instance type, got nil")
		}

		if instanceType.InstanceType != "1V100.6V" {
			t.Errorf("expected instance type 1V100.6V, got %s", instanceType.InstanceType)
		}

		if instanceType.Model != "V100" {
			t.Errorf("expected model V100, got %s", instanceType.Model)
		}
	})

	t.Run("get instance type with all parameters", func(t *testing.T) {
		ctx := context.Background()
		instanceType, err := client.InstanceTypes.GetByInstanceType(ctx, "1H100.80S.22V", true, "FIN-01", "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instanceType == nil {
			t.Fatal("expected instance type, got nil")
		}

		// Verify basic fields
		if instanceType.ID == "" {
			t.Error("expected instance type to have an ID")
		}
		if instanceType.Currency == "" {
			t.Error("expected instance type to have a Currency")
		}
	})

	t.Run("verify pricing fields", func(t *testing.T) {
		ctx := context.Background()
		instanceType, err := client.InstanceTypes.GetByInstanceType(ctx, "1H100.80S.22V", false, "", "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instanceType == nil {
			t.Fatal("expected instance type, got nil")
		}

		// Verify FlexibleFloat pricing fields
		if instanceType.PricePerHour.Float64() <= 0 {
			t.Error("expected price per hour to be positive")
		}
		if instanceType.SpotPrice.Float64() <= 0 {
			t.Error("expected spot price to be positive")
		}
		if instanceType.DynamicPrice.Float64() <= 0 {
			t.Error("expected dynamic price to be positive")
		}
		if instanceType.MaxDynamicPrice.Float64() <= 0 {
			t.Error("expected max dynamic price to be positive")
		}
	})

	t.Run("verify hardware specs", func(t *testing.T) {
		ctx := context.Background()
		instanceType, err := client.InstanceTypes.GetByInstanceType(ctx, "1V100.6V", false, "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instanceType == nil {
			t.Fatal("expected instance type, got nil")
		}

		// Check CPU, GPU, Memory specs
		if instanceType.CPU.NumberOfCores <= 0 {
			t.Error("expected CPU to have number of cores")
		}
		if instanceType.GPU.NumberOfGPUs <= 0 {
			t.Error("expected GPU to have number of GPUs")
		}
		if instanceType.Memory.SizeInGigabytes <= 0 {
			t.Error("expected Memory to have size in gigabytes")
		}
		if instanceType.GPUMemory.SizeInGigabytes <= 0 {
			t.Error("expected GPU Memory to have size in gigabytes")
		}
	})
}

func TestInstanceTypesService_GetPriceHistory(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get price history with default parameters", func(t *testing.T) {
		ctx := context.Background()
		priceHistory, err := client.InstanceTypes.GetPriceHistory(ctx, 0, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(priceHistory) == 0 {
			t.Error("expected price history data")
		}

		// Check that we have H100 and V100 data
		if _, ok := priceHistory["1H100.80S.22V"]; !ok {
			t.Error("expected H100 price history")
		}
		if _, ok := priceHistory["1V100.6V"]; !ok {
			t.Error("expected V100 price history")
		}
	})

	t.Run("get price history for 1 month", func(t *testing.T) {
		ctx := context.Background()
		priceHistory, err := client.InstanceTypes.GetPriceHistory(ctx, 1, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(priceHistory) == 0 {
			t.Error("expected price history data")
		}
	})

	t.Run("get price history for 6 months", func(t *testing.T) {
		ctx := context.Background()
		priceHistory, err := client.InstanceTypes.GetPriceHistory(ctx, 6, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(priceHistory) == 0 {
			t.Error("expected price history data")
		}
	})

	t.Run("verify price history record structure", func(t *testing.T) {
		ctx := context.Background()
		priceHistory, err := client.InstanceTypes.GetPriceHistory(ctx, 1, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check H100 records
		if h100Records, ok := priceHistory["1H100.80S.22V"]; ok && len(h100Records) > 0 {
			record := h100Records[0]

			if record.Date == "" {
				t.Error("expected price record to have a Date")
			}
			if record.Currency == "" {
				t.Error("expected price record to have a Currency")
			}
			if record.FixedPricePerHour.Float64() <= 0 {
				t.Error("expected fixed price per hour to be positive")
			}
			if record.DynamicPricePerHour.Float64() <= 0 {
				t.Error("expected dynamic price per hour to be positive")
			}
		}
	})

	t.Run("verify multiple instance types in history", func(t *testing.T) {
		ctx := context.Background()
		priceHistory, err := client.InstanceTypes.GetPriceHistory(ctx, 1, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(priceHistory) < 2 {
			t.Errorf("expected at least 2 instance types in price history, got %d", len(priceHistory))
		}

		// Verify each instance type has records
		for instanceType, records := range priceHistory {
			if len(records) == 0 {
				t.Errorf("instance type %s has no price records", instanceType)
			}
		}
	})
}
