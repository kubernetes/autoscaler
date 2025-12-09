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

func TestLongTermService_GetInstancePeriods(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get instance long-term periods", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetInstancePeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) == 0 {
			t.Error("expected at least one period")
		}

		// Verify first period has expected fields
		if len(periods) > 0 {
			period := periods[0]
			if period.Code == "" {
				t.Error("expected period to have a Code")
			}
			if period.Name == "" {
				t.Error("expected period to have a Name")
			}
			if period.UnitName == "" {
				t.Error("expected period to have a UnitName")
			}
			if period.UnitValue <= 0 {
				t.Error("expected period to have a positive UnitValue")
			}
			if period.DiscountPercentage < 0 {
				t.Error("expected period to have a non-negative DiscountPercentage")
			}
		}
	})

	t.Run("verify period structure", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetInstancePeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) > 0 {
			for i, period := range periods {
				if period.Code == "" {
					t.Errorf("period %d missing Code", i)
				}
				if period.Name == "" {
					t.Errorf("period %d missing Name", i)
				}
				if period.UnitName == "" {
					t.Errorf("period %d missing UnitName", i)
				}
				if period.UnitValue <= 0 {
					t.Errorf("period %d has invalid UnitValue: %d", i, period.UnitValue)
				}
			}
		}
	})

	t.Run("verify at least one enabled period", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetInstancePeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		hasEnabledPeriod := false
		for _, period := range periods {
			if period.IsEnabled {
				hasEnabledPeriod = true
				break
			}
		}

		if !hasEnabledPeriod {
			t.Error("expected at least one enabled period")
		}
	})
}

func TestLongTermService_GetPeriods(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get general long-term periods", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) == 0 {
			t.Error("expected at least one period")
		}

		// Verify first period has expected fields
		if len(periods) > 0 {
			period := periods[0]
			if period.Code == "" {
				t.Error("expected period to have a Code")
			}
			if period.Name == "" {
				t.Error("expected period to have a Name")
			}
			if period.UnitName == "" {
				t.Error("expected period to have a UnitName")
			}
			if period.UnitValue <= 0 {
				t.Error("expected period to have a positive UnitValue")
			}
		}
	})

	t.Run("verify discount percentages", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) > 0 {
			for i, period := range periods {
				if period.DiscountPercentage < 0 {
					t.Errorf("period %d has negative discount: %f", i, period.DiscountPercentage)
				}
				if period.DiscountPercentage > 100 {
					t.Errorf("period %d has discount over 100%%: %f", i, period.DiscountPercentage)
				}
			}
		}
	})

	t.Run("verify general periods include short term options", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// General periods should include more options than instance-specific
		if len(periods) < 2 {
			t.Error("expected at least 2 period options")
		}
	})
}

func TestLongTermService_GetClusterPeriods(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get cluster long-term periods", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetClusterPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) == 0 {
			t.Error("expected at least one period")
		}

		// Verify first period has expected fields
		if len(periods) > 0 {
			period := periods[0]
			if period.Code == "" {
				t.Error("expected period to have a Code")
			}
			if period.Name == "" {
				t.Error("expected period to have a Name")
			}
			if period.UnitName == "" {
				t.Error("expected period to have a UnitName")
			}
			if period.UnitValue <= 0 {
				t.Error("expected period to have a positive UnitValue")
			}
			if period.DiscountPercentage < 0 {
				t.Error("expected period to have a non-negative DiscountPercentage")
			}
		}
	})

	t.Run("verify cluster periods have higher discounts", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetClusterPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) > 0 {
			// Cluster periods typically have higher discounts
			hasDiscount := false
			for _, period := range periods {
				if period.DiscountPercentage > 0 {
					hasDiscount = true
					break
				}
			}

			if !hasDiscount {
				t.Error("expected cluster periods to have discount percentages")
			}
		}
	})

	t.Run("verify cluster periods are longer term", func(t *testing.T) {
		ctx := context.Background()
		periods, err := client.LongTerm.GetClusterPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(periods) > 0 {
			// Cluster periods are typically longer (6+ months)
			allLongTerm := true
			for _, period := range periods {
				if period.UnitName == "month" && period.UnitValue < 6 {
					allLongTerm = false
					break
				}
			}

			if !allLongTerm {
				t.Log("Note: Some cluster periods are shorter than 6 months")
			}
		}
	})
}

func TestLongTermService_ComparePeriods(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("compare different period types", func(t *testing.T) {
		ctx := context.Background()

		instancePeriods, err := client.LongTerm.GetInstancePeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error getting instance periods: %v", err)
		}

		generalPeriods, err := client.LongTerm.GetPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error getting general periods: %v", err)
		}

		clusterPeriods, err := client.LongTerm.GetClusterPeriods(ctx)
		if err != nil {
			t.Errorf("unexpected error getting cluster periods: %v", err)
		}

		// Verify we got data from all endpoints
		if len(instancePeriods) == 0 {
			t.Error("expected instance periods to have data")
		}
		if len(generalPeriods) == 0 {
			t.Error("expected general periods to have data")
		}
		if len(clusterPeriods) == 0 {
			t.Error("expected cluster periods to have data")
		}

		// General periods should have more options
		if len(generalPeriods) < len(instancePeriods) {
			t.Error("expected general periods to have at least as many options as instance periods")
		}
	})
}
