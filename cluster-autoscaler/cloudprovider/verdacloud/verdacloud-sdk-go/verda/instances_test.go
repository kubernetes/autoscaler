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

func TestInstanceService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all instances", func(t *testing.T) {
		ctx := context.Background()
		instances, err := client.Instances.Get(ctx, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instances) != 1 {
			t.Errorf("expected 1 instance, got %d", len(instances))
		}

		instance := instances[0]
		if instance.ID != "inst_123" {
			t.Errorf("expected instance ID 'inst_123', got '%s'", instance.ID)
		}

		if instance.InstanceType != "1V100.6V" {
			t.Errorf("expected instance type '1V100.6V', got '%s'", instance.InstanceType)
		}

		if instance.Status != StatusRunning {
			t.Errorf("expected status '%s', got '%s'", StatusRunning, instance.Status)
		}
	})

	t.Run("get instances with status filter", func(t *testing.T) {
		ctx := context.Background()
		instances, err := client.Instances.Get(ctx, StatusRunning)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(instances) != 1 {
			t.Errorf("expected 1 instance, got %d", len(instances))
		}
	})
}

func TestInstanceService_GetByID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get instance by ID", func(t *testing.T) {
		ctx := context.Background()
		instanceID := "test_instance_123"
		instance, err := client.Instances.GetByID(ctx, instanceID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instance == nil {
			t.Fatal("expected instance, got nil")
		}

		if instance.ID != instanceID {
			t.Errorf("expected instance ID '%s', got '%s'", instanceID, instance.ID)
		}

		if instance.InstanceType != "1V100.6V" {
			t.Errorf("expected instance type '1V100.6V', got '%s'", instance.InstanceType)
		}
	})
}

func TestInstanceService_Create(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("create instance with minimal config", func(t *testing.T) {
		req := CreateInstanceRequest{
			InstanceType: "1V100.6V",
			Image:        "ubuntu-24.04-cuda-12.8-open-docker",
			Hostname:     "test-instance",
			Description:  "Test instance",
		}

		ctx := context.Background()
		instance, err := client.Instances.Create(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instance == nil {
			t.Fatal("expected instance, got nil")
		}

		if instance.ID != "inst_new_123" {
			t.Errorf("expected instance ID 'inst_new_123', got '%s'", instance.ID)
		}

		if instance.InstanceType != req.InstanceType {
			t.Errorf("expected instance type '%s', got '%s'", req.InstanceType, instance.InstanceType)
		}

		if instance.Hostname != req.Hostname {
			t.Errorf("expected hostname '%s', got '%s'", req.Hostname, instance.Hostname)
		}

		// Should set default location
		if instance.Location != LocationFIN01 {
			t.Errorf("expected location '%s', got '%s'", LocationFIN01, instance.Location)
		}
	})

	t.Run("create instance with full config", func(t *testing.T) {
		req := CreateInstanceRequest{
			InstanceType:    "8V100.48V",
			Image:           "custom-image",
			Hostname:        "ml-server",
			Description:     "ML training server",
			SSHKeyIDs:       []string{"key_123", "key_456"},
			LocationCode:    "US-01",
			Contract:        "PAY_AS_YOU_GO",
			Pricing:         "FIXED_PRICE",
			StartupScriptID: stringPtr("script_123"),
			Volumes: []VolumeCreateRequest{
				{Size: 500, Type: VolumeTypeNVMe, Name: "data"},
			},
			ExistingVolumes: []string{"vol_123"},
			OSVolume:        &OSVolumeCreateRequest{Size: 100},
			IsSpot:          true,
			Coupon:          stringPtr("DISCOUNT20"),
		}

		ctx := context.Background()
		instance, err := client.Instances.Create(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if instance == nil {
			t.Fatal("expected instance, got nil")
		}

		if instance.Location != "US-01" {
			t.Errorf("expected location 'US-01', got '%s'", instance.Location)
		}

		if len(instance.SSHKeyIDs) != 2 {
			t.Errorf("expected 2 SSH keys, got %d", len(instance.SSHKeyIDs))
		}
	})
}

func TestInstanceService_Action(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("action on instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Action(ctx, "inst_123", ActionShutdown, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("action delete with volumes", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Action(ctx, "inst_123", ActionDelete, []string{"vol_123"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestInstanceService_ConvenienceMethods(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("boot instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Boot(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("start instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Start(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("shutdown instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Shutdown(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("delete instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Delete(ctx, "inst_123", []string{"vol_123"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("discontinue instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Discontinue(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("hibernate instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Hibernate(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("configure spot instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.ConfigureSpot(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("force shutdown instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.ForceShutdown(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("delete stuck instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.DeleteStuck(ctx, "inst_123", []string{"vol_123"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("deploy instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Deploy(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("transfer instance", func(t *testing.T) {
		ctx := context.Background()
		err := client.Instances.Transfer(ctx, "inst_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
