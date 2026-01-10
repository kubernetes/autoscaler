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

// Test Balance Service
func TestBalanceService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get balance", func(t *testing.T) {
		ctx := context.Background()
		balance, err := client.Balance.Get(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if balance == nil {
			t.Fatal("expected balance, got nil")
		}

		if balance.Amount != 100.50 {
			t.Errorf("expected amount 100.50, got %f", balance.Amount)
		}

		if balance.Currency != "USD" {
			t.Errorf("expected currency USD, got %s", balance.Currency)
		}
	})
}

// Test SSH Keys Service
func TestSSHKeyService_GetAllSSHKeysFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all SSH keys", func(t *testing.T) {
		ctx := context.Background()
		keys, err := client.SSHKeys.GetAllSSHKeys(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(keys) != 1 {
			t.Errorf("expected 1 SSH key, got %d", len(keys))
		}

		key := keys[0]
		if key.ID != "key_123" {
			t.Errorf("expected key ID 'key_123', got '%s'", key.ID)
		}

		if key.Name != "Test Key" {
			t.Errorf("expected key name 'Test Key', got '%s'", key.Name)
		}

		if key.PublicKey != "ssh-rsa AAAAB3NzaC1yc2E..." {
			t.Errorf("expected public key to be set correctly")
		}
	})
}

func TestSSHKeyService_GetSSHKeyByIDFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for specific SSH key (both old and new paths)
	mockServerHandler := func(w http.ResponseWriter, r *http.Request) {
		key := testutil.SSHKey{
			ID:          "key_123",
			Name:        "Specific Test Key",
			PublicKey:   "ssh-rsa AAAAB3NzaC1yc2E...",
			Fingerprint: "SHA256:abc123...",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]testutil.SSHKey{key})
	}
	mockServer.SetHandler(http.MethodGet, "/ssh-keys/key_123", mockServerHandler)
	mockServer.SetHandler(http.MethodGet, "/sshkeys/key_123", mockServerHandler)

	t.Run("get SSH key by ID", func(t *testing.T) {
		ctx := context.Background()
		key, err := client.SSHKeys.GetSSHKeyByID(ctx, "key_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected SSH key, got nil")
		}

		if key.ID != "key_123" {
			t.Errorf("expected key ID 'key_123', got '%s'", key.ID)
		}

		if key.Name != "Specific Test Key" {
			t.Errorf("expected key name 'Specific Test Key', got '%s'", key.Name)
		}
	})
}

func TestSSHKeyService_AddSSHKeyFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for SSH key creation (both old and new paths)
	createHandler := func(w http.ResponseWriter, r *http.Request) {
		var req CreateSSHKeyRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		key := testutil.SSHKey{
			ID:          "key_new_123",
			Name:        req.Name,
			PublicKey:   req.PublicKey,
			Fingerprint: "SHA256:generated...",
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(key.ID))
	}
	mockServer.SetHandler(http.MethodPost, "/ssh-keys", createHandler)
	mockServer.SetHandler(http.MethodPost, "/sshkeys", createHandler)

	t.Run("create SSH key", func(t *testing.T) {
		req := &CreateSSHKeyRequest{
			Name:      "My New Key",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB...",
		}

		ctx := context.Background()
		key, err := client.SSHKeys.AddSSHKey(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected SSH key, got nil")
		}

		if key.Name != req.Name {
			t.Errorf("expected key name '%s', got '%s'", req.Name, key.Name)
		}

		if key.PublicKey != req.PublicKey {
			t.Errorf("expected public key to match request")
		}
	})
}

func TestSSHKeyService_DeleteSSHKeyFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for SSH key deletion
	mockServer.SetHandler(http.MethodDelete, "/ssh-keys/key_123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	t.Run("delete SSH key", func(t *testing.T) {
		ctx := context.Background()
		err := client.SSHKeys.DeleteSSHKey(ctx, "key_123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Test Locations Service
func TestLocationService_Get(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all locations", func(t *testing.T) {
		ctx := context.Background()
		locations, err := client.Locations.Get(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(locations) != 1 {
			t.Errorf("expected 1 location, got %d", len(locations))
		}

		location := locations[0]
		if location.Code != LocationFIN01 {
			t.Errorf("expected location code '%s', got '%s'", LocationFIN01, location.Code)
		}

		if location.Name != "Finland 01" {
			t.Errorf("expected location name 'Finland 01', got '%s'", location.Name)
		}

		if location.CountryCode != "FI" {
			t.Errorf("expected country code 'FI', got '%s'", location.CountryCode)
		}
	})
}

// Test Volumes Service
func TestVolumeService_ListVolumesFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for volumes
	mockServer.SetHandler(http.MethodGet, "/volumes", func(w http.ResponseWriter, r *http.Request) {
		volumes := []testutil.Volume{
			{
				ID:     "vol_123",
				Name:   "Test Volume",
				Size:   100,
				Type:   VolumeTypeNVMe,
				Status: "available",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(volumes)
	})

	t.Run("get all volumes", func(t *testing.T) {
		ctx := context.Background()
		volumes, err := client.Volumes.ListVolumes(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(volumes) != 1 {
			t.Errorf("expected 1 volume, got %d", len(volumes))
		}

		volume := volumes[0]
		if volume.ID != "vol_123" {
			t.Errorf("expected volume ID 'vol_123', got '%s'", volume.ID)
		}

		if volume.Size != 100 {
			t.Errorf("expected volume size 100, got %d", volume.Size)
		}
	})
}

func TestVolumeService_GetVolumeFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for specific volume
	mockServer.SetHandler(http.MethodGet, "/volumes/vol_123", func(w http.ResponseWriter, r *http.Request) {
		volume := testutil.Volume{
			ID:     "vol_123",
			Name:   "Specific Volume",
			Size:   200,
			Type:   VolumeTypeNVMe,
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
	})
}

// Test Startup Scripts Service
func TestStartupScriptService_GetAllStartupScriptsFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for startup scripts
	mockServer.SetHandler(http.MethodGet, "/scripts", func(w http.ResponseWriter, r *http.Request) {
		scripts := []testutil.StartupScript{
			{
				ID:     "script_123",
				Name:   "Test Script",
				Script: "#!/bin/bash\necho 'Hello World'",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(scripts)
	})

	t.Run("get all startup scripts", func(t *testing.T) {
		ctx := context.Background()
		scripts, err := client.StartupScripts.GetAllStartupScripts(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(scripts) != 1 {
			t.Errorf("expected 1 startup script, got %d", len(scripts))
		}

		script := scripts[0]
		if script.ID != "script_123" {
			t.Errorf("expected script ID 'script_123', got '%s'", script.ID)
		}

		if script.Name != "Test Script" {
			t.Errorf("expected script name 'Test Script', got '%s'", script.Name)
		}
	})
}

func TestStartupScriptService_AddStartupScriptFromServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Set up mock response for startup script creation (returns plain text ID)
	mockServer.SetHandler(http.MethodPost, "/scripts", func(w http.ResponseWriter, r *http.Request) {
		var req CreateStartupScriptRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Return plain text ID (matching real API behavior)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("script_new_123"))
	})

	// Set up mock response for GetByID (returns array)
	mockServer.SetHandler(http.MethodGet, "/scripts/script_new_123", func(w http.ResponseWriter, r *http.Request) {
		scripts := []testutil.StartupScript{
			{
				ID:     "script_new_123",
				Name:   "Setup Script",
				Script: "#!/bin/bash\nnpm install",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(scripts)
	})

	t.Run("create startup script", func(t *testing.T) {
		req := &CreateStartupScriptRequest{
			Name:   "Setup Script",
			Script: "#!/bin/bash\nnpm install",
		}

		ctx := context.Background()
		script, err := client.StartupScripts.AddStartupScript(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if script == nil {
			t.Fatal("expected startup script, got nil")
		}

		if script.Name != req.Name {
			t.Errorf("expected script name '%s', got '%s'", req.Name, script.Name)
		}

		if script.Script != req.Script {
			t.Errorf("expected script content to match request")
		}
	})
}
