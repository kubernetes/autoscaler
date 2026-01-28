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

func TestInstanceAvailabilityService_GetAllAvailabilities(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all availabilities with default parameters", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.InstanceAvailability.GetAllAvailabilities(ctx, false, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) == 0 {
			t.Error("expected at least one availability entry")
		}

		// Verify first availability has expected fields
		if len(availabilities) > 0 {
			avail := availabilities[0]
			if avail.LocationCode == "" {
				t.Error("expected availability to have a LocationCode")
			}
			if avail.Availabilities == nil {
				t.Error("expected availability to have an Availabilities slice")
			}
		}
	})

	t.Run("get availabilities for specific location", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.InstanceAvailability.GetAllAvailabilities(ctx, false, "FIN-01")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) == 0 {
			t.Error("expected at least one availability entry")
		}
	})

	t.Run("get spot instance availabilities", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.InstanceAvailability.GetAllAvailabilities(ctx, true, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) == 0 {
			t.Error("expected at least one availability entry")
		}
	})

	t.Run("get spot availabilities for specific location", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.InstanceAvailability.GetAllAvailabilities(ctx, true, "FIN-01")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) == 0 {
			t.Error("expected at least one availability entry")
		}
	})

	t.Run("verify availabilities structure", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.InstanceAvailability.GetAllAvailabilities(ctx, false, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) > 0 {
			for i, avail := range availabilities {
				if avail.LocationCode == "" {
					t.Errorf("availability %d missing LocationCode", i)
				}
				if avail.Availabilities == nil {
					t.Errorf("availability %d has nil Availabilities field", i)
				}
				// Check that at least some locations have available instance types
				if len(avail.Availabilities) == 0 {
					t.Logf("availability %d for location %s has no available instance types", i, avail.LocationCode)
				}
			}
		}
	})
}

func TestInstanceAvailabilityService_GetInstanceTypeAvailability(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("check availability for H100 instance type", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.InstanceAvailability.GetInstanceTypeAvailability(ctx, "1H100.80S.22V", false, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !available {
			t.Error("expected H100 instance type to be available")
		}
	})

	t.Run("check availability for V100 instance type", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.InstanceAvailability.GetInstanceTypeAvailability(ctx, "1V100.6V", false, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !available {
			t.Error("expected V100 instance type to be available")
		}
	})

	t.Run("check spot instance availability", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.InstanceAvailability.GetInstanceTypeAvailability(ctx, "1H100.80S.22V", true, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Availability is boolean, just checking we get a valid response
		_ = available
	})

	t.Run("check availability for specific location", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.InstanceAvailability.GetInstanceTypeAvailability(ctx, "1V100.6V", false, "FIN-01")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Availability is boolean, just checking we get a valid response
		_ = available
	})

	t.Run("check availability with all parameters", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.InstanceAvailability.GetInstanceTypeAvailability(ctx, "1H100.80S.22V", true, "FIN-01")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Availability is boolean, just checking we get a valid response
		_ = available
	})
}
