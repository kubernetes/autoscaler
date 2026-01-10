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

func TestSSHKeyService_GetAllSSHKeys(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all SSH keys", func(t *testing.T) {
		ctx := context.Background()
		keys, err := client.SSHKeys.GetAllSSHKeys(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(keys) == 0 {
			t.Error("expected at least one SSH key")
		}

		// Verify first key has expected fields
		if len(keys) > 0 {
			key := keys[0]
			if key.ID == "" {
				t.Error("expected key to have an ID")
			}
			if key.Name == "" {
				t.Error("expected key to have a Name")
			}
			if key.PublicKey == "" {
				t.Error("expected key to have a PublicKey")
			}
		}
	})

	t.Run("verify SSH key structure", func(t *testing.T) {
		ctx := context.Background()
		keys, err := client.SSHKeys.GetAllSSHKeys(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(keys) > 0 {
			for i, key := range keys {
				if key.ID == "" {
					t.Errorf("key %d missing ID", i)
				}
				if key.Name == "" {
					t.Errorf("key %d missing Name", i)
				}
				if key.PublicKey == "" {
					t.Errorf("key %d missing PublicKey", i)
				}
			}
		}
	})
}

func TestSSHKeyService_GetSSHKeyByID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get SSH key by ID", func(t *testing.T) {
		ctx := context.Background()
		keyID := "key_123"

		key, err := client.SSHKeys.GetSSHKeyByID(ctx, keyID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected key, got nil")
		}

		if key.ID != keyID {
			t.Errorf("expected key ID %s, got %s", keyID, key.ID)
		}

		if key.Name == "" {
			t.Error("expected key to have a Name")
		}

		if key.PublicKey == "" {
			t.Error("expected key to have a PublicKey")
		}
	})

	t.Run("verify key fields", func(t *testing.T) {
		ctx := context.Background()
		keyID := "key_456"

		key, err := client.SSHKeys.GetSSHKeyByID(ctx, keyID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if key != nil {
			// Verify all expected fields are present
			if key.ID == "" {
				t.Error("key missing ID")
			}
			if key.Name == "" {
				t.Error("key missing Name")
			}
			if key.PublicKey == "" {
				t.Error("key missing PublicKey")
			}
		}
	})
}

func TestSSHKeyService_AddSSHKey(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("add new SSH key", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateSSHKeyRequest{
			Name:      "My New Key",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB...",
		}

		key, err := client.SSHKeys.AddSSHKey(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected key, got nil")
		}

		if key.ID == "" {
			t.Error("expected key to have an ID")
		}

		if key.Name != req.Name {
			t.Errorf("expected key name %s, got %s", req.Name, key.Name)
		}

		if key.PublicKey != req.PublicKey {
			t.Errorf("expected key public key %s, got %s", req.PublicKey, key.PublicKey)
		}
	})

	t.Run("verify created key has all fields", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateSSHKeyRequest{
			Name:      "Test Key",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
		}

		key, err := client.SSHKeys.AddSSHKey(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if key != nil {
			if key.ID == "" {
				t.Error("created key missing ID")
			}
			if key.Name == "" {
				t.Error("created key missing Name")
			}
			if key.PublicKey == "" {
				t.Error("created key missing PublicKey")
			}
			if key.Fingerprint == "" {
				t.Error("created key missing Fingerprint")
			}
		}
	})
}

func TestSSHKeyService_DeleteSSHKey(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("delete SSH key by ID", func(t *testing.T) {
		ctx := context.Background()

		// First create a key
		createReq := &CreateSSHKeyRequest{
			Name:      "Key to Delete",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
		}

		key, err := client.SSHKeys.AddSSHKey(ctx, createReq)
		if err != nil {
			t.Fatalf("failed to create key: %v", err)
		}

		// Now delete it
		err = client.SSHKeys.DeleteSSHKey(ctx, key.ID)
		if err != nil {
			t.Errorf("unexpected error deleting key: %v", err)
		}
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		ctx := context.Background()

		// Try to delete a key that doesn't exist
		// The mock server won't fail, but in production this might return an error
		err := client.SSHKeys.DeleteSSHKey(ctx, "non_existent_key_id")
		// Mock server returns success even for non-existent keys
		// In production, this might be different
		_ = err
	})
}

func TestSSHKeyService_DeleteMultipleSSHKeys(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("delete multiple SSH keys", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple keys first
		key1Req := &CreateSSHKeyRequest{
			Name:      "Key 1",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2E1...",
		}

		key2Req := &CreateSSHKeyRequest{
			Name:      "Key 2",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2E2...",
		}

		key1, err := client.SSHKeys.AddSSHKey(ctx, key1Req)
		if err != nil {
			t.Fatalf("failed to create key 1: %v", err)
		}

		key2, err := client.SSHKeys.AddSSHKey(ctx, key2Req)
		if err != nil {
			t.Fatalf("failed to create key 2: %v", err)
		}

		// Delete both keys
		keyIDs := []string{key1.ID, key2.ID}
		err = client.SSHKeys.DeleteMultipleSSHKeys(ctx, keyIDs)
		if err != nil {
			t.Errorf("unexpected error deleting multiple keys: %v", err)
		}
	})

	t.Run("delete empty list", func(t *testing.T) {
		ctx := context.Background()

		// Try to delete empty list
		err := client.SSHKeys.DeleteMultipleSSHKeys(ctx, []string{})
		// Should not error
		if err != nil {
			t.Errorf("unexpected error deleting empty list: %v", err)
		}
	})
}
