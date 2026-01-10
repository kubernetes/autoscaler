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
	"encoding/json"
	"net/http"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

func TestVolumeService_ListVolumes(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volumes
	mockServer.SetHandler(http.MethodGet, "/volumes", func(w http.ResponseWriter, r *http.Request) {
		volumes := []testutil.Volume{
			{
				ID:     "vol_123",
				Name:   "Test Volume 1",
				Size:   100,
				Type:   "NVMe",
				Status: "available",
			},
			{
				ID:     "vol_456",
				Name:   "Test Volume 2",
				Size:   200,
				Type:   "HDD",
				Status: "attached",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(volumes)
	})

	t.Run("list all volumes", func(t *testing.T) {
		ctx := context.Background()
		volumes, err := client.Volumes.ListVolumes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumes) != 2 {
			t.Errorf("expected 2 volumes, got %d", len(volumes))
		}

		if volumes[0].ID != "vol_123" {
			t.Errorf("expected volume ID 'vol_123', got '%s'", volumes[0].ID)
		}

		if volumes[1].Size != 200 {
			t.Errorf("expected volume size 200, got %d", volumes[1].Size)
		}
	})
}

func TestVolumeService_ListVolumesByStatus(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volumes by status
	mockServer.SetHandler(http.MethodGet, "/volumes", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		volumes := []testutil.Volume{}

		switch status {
		case "attached":
			volumes = append(volumes, testutil.Volume{
				ID:     "vol_attached_1",
				Name:   "Attached Volume",
				Size:   100,
				Type:   "NVMe",
				Status: "attached",
			})
		case "detached":
			volumes = append(volumes, testutil.Volume{
				ID:     "vol_detached_1",
				Name:   "Detached Volume",
				Size:   150,
				Type:   "HDD",
				Status: "detached",
			})
		default:
			// Return all volumes
			volumes = append(volumes,
				testutil.Volume{ID: "vol_1", Name: "Volume 1", Size: 100, Type: "NVMe", Status: "attached"},
				testutil.Volume{ID: "vol_2", Name: "Volume 2", Size: 200, Type: "HDD", Status: "detached"},
			)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(volumes)
	})

	t.Run("list volumes by status - attached", func(t *testing.T) {
		ctx := context.Background()
		volumes, err := client.Volumes.ListVolumesByStatus(ctx, "attached")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumes) != 1 {
			t.Errorf("expected 1 volume, got %d", len(volumes))
		}

		if volumes[0].Status != "attached" {
			t.Errorf("expected volume status 'attached', got '%s'", volumes[0].Status)
		}
	})

	t.Run("list volumes by status - detached", func(t *testing.T) {
		ctx := context.Background()
		volumes, err := client.Volumes.ListVolumesByStatus(ctx, "detached")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumes) != 1 {
			t.Errorf("expected 1 volume, got %d", len(volumes))
		}

		if volumes[0].Status != "detached" {
			t.Errorf("expected volume status 'detached', got '%s'", volumes[0].Status)
		}
	})
}

func TestVolumeService_GetVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for specific volume
	mockServer.SetHandler(http.MethodGet, "/volumes/vol_123", func(w http.ResponseWriter, r *http.Request) {
		volume := testutil.Volume{
			ID:     "vol_123",
			Name:   "Specific Volume",
			Size:   200,
			Type:   "NVMe",
			Status: "in-use",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(volume)
	})

	t.Run("get volume by ID", func(t *testing.T) {
		ctx := context.Background()
		volume, err := client.Volumes.GetVolume(ctx, "vol_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if volume == nil {
			t.Fatal("expected volume, got nil")
		}

		if volume.ID != "vol_123" {
			t.Errorf("expected volume ID 'vol_123', got '%s'", volume.ID)
		}

		if volume.Size != 200 {
			t.Errorf("expected volume size 200, got %d", volume.Size)
		}

		if volume.Name != "Specific Volume" {
			t.Errorf("expected volume name 'Specific Volume', got '%s'", volume.Name)
		}
	})
}

func TestVolumeService_CreateVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume creation (returns plain text ID)
	mockServer.SetHandler(http.MethodPost, "/volumes", func(w http.ResponseWriter, r *http.Request) {
		var req VolumeCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Validate request
		if req.Size <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid size"))
			return
		}

		// Return plain text ID (matching real API behavior)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("vol_new_123"))
	})

	t.Run("create volume", func(t *testing.T) {
		req := VolumeCreateRequest{
			Type:         VolumeTypeNVMe,
			LocationCode: LocationFIN01,
			Size:         100,
			Name:         "New Test Volume",
		}

		ctx := context.Background()
		volumeID, err := client.Volumes.CreateVolume(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if volumeID != "vol_new_123" {
			t.Errorf("expected volume ID 'vol_new_123', got '%s'", volumeID)
		}
	})

	t.Run("create volume with default location", func(t *testing.T) {
		req := VolumeCreateRequest{
			Type: VolumeTypeHDD,
			Size: 200,
			Name: "Volume Without Location",
		}

		ctx := context.Background()
		volumeID, err := client.Volumes.CreateVolume(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if volumeID == "" {
			t.Error("expected volume ID, got empty string")
		}
	})
}

func TestVolumeService_DeleteVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume deletion
	mockServer.SetHandler(http.MethodDelete, "/volumes/vol_123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	t.Run("delete volume", func(t *testing.T) {
		ctx := context.Background()
		err := client.Volumes.DeleteVolume(ctx, "vol_123", false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("delete volume with force", func(t *testing.T) {
		ctx := context.Background()
		err := client.Volumes.DeleteVolume(ctx, "vol_123", true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestVolumeService_AttachVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume attach
	mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/attach", func(w http.ResponseWriter, r *http.Request) {
		var req VolumeAttachRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.InstanceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("instance_id required"))
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	t.Run("attach volume to instance", func(t *testing.T) {
		req := VolumeAttachRequest{
			InstanceID: "inst_123",
		}

		ctx := context.Background()
		err := client.Volumes.AttachVolume(ctx, "vol_123", req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestVolumeService_DetachVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume detach
	mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/detach", func(w http.ResponseWriter, r *http.Request) {
		var req VolumeDetachRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.InstanceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("instance_id required"))
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	t.Run("detach volume from instance", func(t *testing.T) {
		req := VolumeDetachRequest{
			InstanceID: "inst_123",
		}

		ctx := context.Background()
		err := client.Volumes.DetachVolume(ctx, "vol_123", req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestVolumeService_CloneVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume clone
	mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/clone", func(w http.ResponseWriter, r *http.Request) {
		var req VolumeCloneRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("name required"))
			return
		}

		// Return plain text ID (matching real API behavior)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("vol_cloned_456"))
	})

	t.Run("clone volume", func(t *testing.T) {
		req := VolumeCloneRequest{
			Name:         "Cloned Volume",
			LocationCode: LocationFIN01,
		}

		ctx := context.Background()
		newVolumeID, err := client.Volumes.CloneVolume(ctx, "vol_123", req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if newVolumeID != "vol_cloned_456" {
			t.Errorf("expected volume ID 'vol_cloned_456', got '%s'", newVolumeID)
		}
	})
}

func TestVolumeService_ResizeVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume resize
	mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/resize", func(w http.ResponseWriter, r *http.Request) {
		var req VolumeResizeRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Size <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid size"))
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	t.Run("resize volume", func(t *testing.T) {
		req := VolumeResizeRequest{
			Size: 200,
		}

		ctx := context.Background()
		err := client.Volumes.ResizeVolume(ctx, "vol_123", req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestVolumeService_RenameVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volume rename
	mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/rename", func(w http.ResponseWriter, r *http.Request) {
		var req VolumeRenameRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("name required"))
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	t.Run("rename volume", func(t *testing.T) {
		req := VolumeRenameRequest{
			Name: "New Volume Name",
		}

		ctx := context.Background()
		err := client.Volumes.RenameVolume(ctx, "vol_123", req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Test error scenarios
func TestVolumeService_ErrorHandling(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get non-existent volume", func(t *testing.T) {
		mockServer.SetHandler(http.MethodGet, "/volumes/vol_nonexistent", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "Volume not found"}`))
		})

		ctx := context.Background()
		_, err := client.Volumes.GetVolume(ctx, "vol_nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("create volume with invalid size", func(t *testing.T) {
		mockServer.SetHandler(http.MethodPost, "/volumes", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Invalid volume size"}`))
		})

		req := VolumeCreateRequest{
			Type:         VolumeTypeNVMe,
			LocationCode: LocationFIN01,
			Size:         0, // Invalid size
			Name:         "Invalid Volume",
		}

		ctx := context.Background()
		_, err := client.Volumes.CreateVolume(ctx, req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("attach volume to non-existent instance", func(t *testing.T) {
		mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/attach", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "Instance not found"}`))
		})

		req := VolumeAttachRequest{
			InstanceID: "nonexistent",
		}

		ctx := context.Background()
		err := client.Volumes.AttachVolume(ctx, "vol_123", req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("resize volume to smaller size", func(t *testing.T) {
		mockServer.SetHandler(http.MethodPost, "/volumes/vol_123/resize", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Cannot shrink volume"}`))
		})

		req := VolumeResizeRequest{
			Size: 50, // Smaller than current
		}

		ctx := context.Background()
		err := client.Volumes.ResizeVolume(ctx, "vol_123", req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
