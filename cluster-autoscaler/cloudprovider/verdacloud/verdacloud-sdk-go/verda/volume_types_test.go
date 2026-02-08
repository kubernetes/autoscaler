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

func TestVolumeTypeService_GetAllVolumeTypes(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all volume types", func(t *testing.T) {
		ctx := context.Background()
		volumeTypes, err := client.VolumeTypes.GetAllVolumeTypes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumeTypes) == 0 {
			t.Error("expected at least one volume type")
		}

		// Verify first volume type has expected fields
		if len(volumeTypes) > 0 {
			vt := volumeTypes[0]
			if vt.Type == "" {
				t.Error("expected volume type to have a Type")
			}
			if vt.Price.MonthlyPerGB <= 0 {
				t.Error("expected volume type to have a positive price per GB")
			}
			if vt.Price.Currency == "" {
				t.Error("expected volume type to have a Currency")
			}
			if vt.IOPS == "" {
				t.Error("expected volume type to have IOPS")
			}
		}
	})

	t.Run("verify volume type structure", func(t *testing.T) {
		ctx := context.Background()
		volumeTypes, err := client.VolumeTypes.GetAllVolumeTypes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumeTypes) > 0 {
			for i, vt := range volumeTypes {
				if vt.Type == "" {
					t.Errorf("volume type %d missing Type", i)
				}
				if vt.Price.MonthlyPerGB <= 0 {
					t.Errorf("volume type %d has invalid price: %f", i, vt.Price.MonthlyPerGB)
				}
				if vt.Price.Currency == "" {
					t.Errorf("volume type %d missing Currency", i)
				}
				if vt.BurstBandwidth < 0 {
					t.Errorf("volume type %d has negative BurstBandwidth: %f", i, vt.BurstBandwidth)
				}
				if vt.ContinuousBandwidth < 0 {
					t.Errorf("volume type %d has negative ContinuousBandwidth: %f", i, vt.ContinuousBandwidth)
				}
				if vt.InternalNetworkSpeed < 0 {
					t.Errorf("volume type %d has negative InternalNetworkSpeed: %f", i, vt.InternalNetworkSpeed)
				}
				if vt.IOPS == "" {
					t.Errorf("volume type %d missing IOPS", i)
				}
			}
		}
	})

	t.Run("verify specific volume types", func(t *testing.T) {
		ctx := context.Background()
		volumeTypes, err := client.VolumeTypes.GetAllVolumeTypes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check that we have common volume types
		foundNVMe := false
		foundHDD := false

		for _, vt := range volumeTypes {
			if vt.Type == "NVMe" {
				foundNVMe = true
				if vt.BurstBandwidth <= 0 {
					t.Error("NVMe volume type should have positive burst bandwidth")
				}
			}
			if vt.Type == "HDD" {
				foundHDD = true
			}
		}

		if !foundNVMe && !foundHDD {
			t.Log("Warning: Expected to find common volume types (NVMe or HDD)")
		}
	})

	t.Run("verify price structure", func(t *testing.T) {
		ctx := context.Background()
		volumeTypes, err := client.VolumeTypes.GetAllVolumeTypes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumeTypes) > 0 {
			for i, vt := range volumeTypes {
				// Verify price object is populated
				if vt.Price.MonthlyPerGB <= 0 {
					t.Errorf("volume type %d (%s) has invalid monthly price: %f",
						i, vt.Type, vt.Price.MonthlyPerGB)
				}
				if vt.Price.Currency != "USD" && vt.Price.Currency != "EUR" && vt.Price.Currency != "" {
					t.Logf("volume type %d (%s) has unusual currency: %s",
						i, vt.Type, vt.Price.Currency)
				}
			}
		}
	})

	t.Run("verify performance metrics", func(t *testing.T) {
		ctx := context.Background()
		volumeTypes, err := client.VolumeTypes.GetAllVolumeTypes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumeTypes) > 0 {
			for i, vt := range volumeTypes {
				// Continuous bandwidth should not exceed burst bandwidth
				if vt.ContinuousBandwidth > vt.BurstBandwidth {
					t.Errorf("volume type %d (%s) continuous bandwidth (%f) exceeds burst bandwidth (%f)",
						i, vt.Type, vt.ContinuousBandwidth, vt.BurstBandwidth)
				}

				// Verify IOPS is a non-empty string
				if vt.IOPS == "" {
					t.Errorf("volume type %d (%s) missing IOPS specification", i, vt.Type)
				}
			}
		}
	})

	t.Run("verify shared filesystem flag", func(t *testing.T) {
		ctx := context.Background()
		volumeTypes, err := client.VolumeTypes.GetAllVolumeTypes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check that the IsSharedFS field is being set (boolean field)
		if len(volumeTypes) > 0 {
			for _, vt := range volumeTypes {
				// Just verify the field exists and can be read
				// (boolean always has a value: true or false)
				_ = vt.IsSharedFS
			}
		}
	})
}
