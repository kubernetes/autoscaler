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

func TestStartupScriptService_GetAllStartupScripts(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get all startup scripts", func(t *testing.T) {
		ctx := context.Background()
		scripts, err := client.StartupScripts.GetAllStartupScripts(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(scripts) == 0 {
			t.Error("expected at least one startup script")
		}

		// Verify first script has expected fields
		if len(scripts) > 0 {
			script := scripts[0]
			if script.ID == "" {
				t.Error("expected script to have an ID")
			}
			if script.Name == "" {
				t.Error("expected script to have a Name")
			}
			if script.Script == "" {
				t.Error("expected script to have Script content")
			}
		}
	})

	t.Run("verify startup script structure", func(t *testing.T) {
		ctx := context.Background()
		scripts, err := client.StartupScripts.GetAllStartupScripts(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(scripts) > 0 {
			for i, script := range scripts {
				if script.ID == "" {
					t.Errorf("script %d missing ID", i)
				}
				if script.Name == "" {
					t.Errorf("script %d missing Name", i)
				}
				if script.Script == "" {
					t.Errorf("script %d missing Script content", i)
				}
			}
		}
	})
}

func TestStartupScriptService_GetStartupScriptByID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get startup script by ID", func(t *testing.T) {
		ctx := context.Background()
		scriptID := "script_123"

		script, err := client.StartupScripts.GetStartupScriptByID(ctx, scriptID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if script == nil {
			t.Fatal("expected script, got nil")
		}

		if script.ID != scriptID {
			t.Errorf("expected script ID %s, got %s", scriptID, script.ID)
		}

		if script.Name == "" {
			t.Error("expected script to have a Name")
		}

		if script.Script == "" {
			t.Error("expected script to have Script content")
		}
	})

	t.Run("verify script fields", func(t *testing.T) {
		ctx := context.Background()
		scriptID := "script_456"

		script, err := client.StartupScripts.GetStartupScriptByID(ctx, scriptID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if script != nil {
			// Verify all expected fields are present
			if script.ID == "" {
				t.Error("script missing ID")
			}
			if script.Name == "" {
				t.Error("script missing Name")
			}
			if script.Script == "" {
				t.Error("script missing Script content")
			}
		}
	})
}

func TestStartupScriptService_AddStartupScript(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("add new startup script", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateStartupScriptRequest{
			Name:   "My Startup Script",
			Script: "#!/bin/bash\necho 'Hello World'",
		}

		script, err := client.StartupScripts.AddStartupScript(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if script == nil {
			t.Fatal("expected script, got nil")
		}

		if script.ID == "" {
			t.Error("expected script to have an ID")
		}

		if script.Name != req.Name {
			t.Errorf("expected script name %s, got %s", req.Name, script.Name)
		}

		if script.Script != req.Script {
			t.Errorf("expected script content %s, got %s", req.Script, script.Script)
		}
	})

	t.Run("verify created script has all fields", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateStartupScriptRequest{
			Name:   "Test Script",
			Script: "#!/bin/sh\napt-get update",
		}

		script, err := client.StartupScripts.AddStartupScript(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if script != nil {
			if script.ID == "" {
				t.Error("created script missing ID")
			}
			if script.Name == "" {
				t.Error("created script missing Name")
			}
			if script.Script == "" {
				t.Error("created script missing Script content")
			}
		}
	})
}

func TestStartupScriptService_DeleteStartupScript(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("delete startup script by ID", func(t *testing.T) {
		ctx := context.Background()

		// First create a script
		createReq := &CreateStartupScriptRequest{
			Name:   "Script to Delete",
			Script: "#!/bin/bash\necho 'test'",
		}

		script, err := client.StartupScripts.AddStartupScript(ctx, createReq)
		if err != nil {
			t.Fatalf("failed to create script: %v", err)
		}

		// Now delete it
		err = client.StartupScripts.DeleteStartupScript(ctx, script.ID)
		if err != nil {
			t.Errorf("unexpected error deleting script: %v", err)
		}
	})

	t.Run("delete non-existent script", func(t *testing.T) {
		ctx := context.Background()

		// Try to delete a script that doesn't exist
		// The mock server won't fail, but in production this might return an error
		err := client.StartupScripts.DeleteStartupScript(ctx, "non_existent_script_id")
		// Mock server returns success even for non-existent scripts
		// In production, this might be different
		_ = err
	})
}

func TestStartupScriptService_DeleteMultipleStartupScripts(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("delete multiple startup scripts", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple scripts first
		script1Req := &CreateStartupScriptRequest{
			Name:   "Script 1",
			Script: "#!/bin/bash\necho 'script1'",
		}

		script2Req := &CreateStartupScriptRequest{
			Name:   "Script 2",
			Script: "#!/bin/bash\necho 'script2'",
		}

		script1, err := client.StartupScripts.AddStartupScript(ctx, script1Req)
		if err != nil {
			t.Fatalf("failed to create script 1: %v", err)
		}

		script2, err := client.StartupScripts.AddStartupScript(ctx, script2Req)
		if err != nil {
			t.Fatalf("failed to create script 2: %v", err)
		}

		// Delete both scripts
		scriptIDs := []string{script1.ID, script2.ID}
		err = client.StartupScripts.DeleteMultipleStartupScripts(ctx, scriptIDs)
		if err != nil {
			t.Errorf("unexpected error deleting multiple scripts: %v", err)
		}
	})

	t.Run("delete empty list", func(t *testing.T) {
		ctx := context.Background()

		// Try to delete empty list
		err := client.StartupScripts.DeleteMultipleStartupScripts(ctx, []string{})
		// Should not error
		if err != nil {
			t.Errorf("unexpected error deleting empty list: %v", err)
		}
	})
}
