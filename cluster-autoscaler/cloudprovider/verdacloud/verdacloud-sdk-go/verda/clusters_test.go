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

func TestClusterService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all clusters", func(t *testing.T) {
		ctx := context.Background()
		clusters, err := client.Clusters.Get(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(clusters) == 0 {
			t.Error("expected at least one cluster")
		}
	})
}

func TestClusterService_GetByID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get cluster by ID", func(t *testing.T) {
		ctx := context.Background()
		clusterID := "test_cluster_123"
		cluster, err := client.Clusters.GetByID(ctx, clusterID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cluster == nil {
			t.Fatal("expected cluster, got nil")
		}

		if cluster.ID != clusterID {
			t.Errorf("expected cluster ID '%s', got '%s'", clusterID, cluster.ID)
		}
	})
}

func TestClusterService_Create(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("create cluster with minimal config", func(t *testing.T) {
		req := CreateClusterRequest{
			ClusterType: "8V100.48V",
			Image:       "ubuntu-22.04-cuda-12.0",
			Hostname:    "test-cluster",
			Description: "Test cluster",
			SSHKeyIDs:   []string{"key_123"},
		}

		ctx := context.Background()
		response, err := client.Clusters.Create(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		if response.ID == "" {
			t.Error("expected cluster ID, got empty string")
		}
	})

	t.Run("create cluster with full config", func(t *testing.T) {
		startupScriptID := "script_123"
		coupon := "TEST_COUPON"

		req := CreateClusterRequest{
			ClusterType:     "8V100.48V",
			Image:           "ubuntu-22.04-cuda-12.0",
			Hostname:        "test-cluster-full",
			Description:     "Test cluster with full config",
			SSHKeyIDs:       []string{"key_123", "key_456"},
			LocationCode:    LocationFIN01,
			Contract:        "hourly",
			Pricing:         "on-demand",
			StartupScriptID: &startupScriptID,
			SharedVolumes:   []string{"vol_123"},
			ExistingVolumes: []string{"vol_456"},
			Coupon:          &coupon,
		}

		ctx := context.Background()
		response, err := client.Clusters.Create(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if response == nil {
			t.Fatal("expected response, got nil")
		}
	})
}

func TestClusterService_Action(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("action on single cluster", func(t *testing.T) {
		ctx := context.Background()
		err := client.Clusters.Action(ctx, "cluster_123", ClusterActionDiscontinue)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("action on multiple clusters", func(t *testing.T) {
		ctx := context.Background()
		clusterIDs := []string{"cluster_123", "cluster_456"}
		err := client.Clusters.Action(ctx, clusterIDs, ClusterActionDiscontinue)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClusterService_ConvenienceMethods(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("discontinue cluster", func(t *testing.T) {
		ctx := context.Background()
		err := client.Clusters.Discontinue(ctx, "cluster_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("discontinue multiple clusters", func(t *testing.T) {
		ctx := context.Background()
		clusterIDs := []string{"cluster_123", "cluster_456"}
		err := client.Clusters.Discontinue(ctx, clusterIDs)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClusterService_GetClusterTypes(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get cluster types with default currency", func(t *testing.T) {
		ctx := context.Background()
		clusterTypes, err := client.Clusters.GetClusterTypes(ctx, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(clusterTypes) == 0 {
			t.Error("expected at least one cluster type")
		}
	})

	t.Run("get cluster types with USD currency", func(t *testing.T) {
		ctx := context.Background()
		clusterTypes, err := client.Clusters.GetClusterTypes(ctx, "usd")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(clusterTypes) == 0 {
			t.Error("expected at least one cluster type")
		}
	})
}

func TestClusterService_GetAvailabilities(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all cluster availabilities", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.Clusters.GetAvailabilities(ctx, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) == 0 {
			t.Error("expected at least one availability entry")
		}
	})

	t.Run("get cluster availabilities for specific location", func(t *testing.T) {
		ctx := context.Background()
		availabilities, err := client.Clusters.GetAvailabilities(ctx, LocationFIN01)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(availabilities) == 0 {
			t.Error("expected at least one availability entry")
		}
	})
}

func TestClusterService_CheckClusterTypeAvailability(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("check cluster type availability", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.Clusters.CheckClusterTypeAvailability(ctx, "8V100.48V", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// We just check that the call succeeded
		_ = available
	})

	t.Run("check cluster type availability for specific location", func(t *testing.T) {
		ctx := context.Background()
		available, err := client.Clusters.CheckClusterTypeAvailability(ctx, "8V100.48V", LocationFIN01)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// We just check that the call succeeded
		_ = available
	})
}

func TestClusterService_GetImages(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get cluster images", func(t *testing.T) {
		ctx := context.Background()
		images, err := client.Clusters.GetImages(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(images) == 0 {
			t.Error("expected at least one cluster image")
		}
	})
}
