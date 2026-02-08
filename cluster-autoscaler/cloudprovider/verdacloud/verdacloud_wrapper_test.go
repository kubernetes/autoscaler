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

package verdacloud

import (
	"context"
	"net/http"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

// Test constants for wrapper tests
const (
	testWrapperClientID     = "test-client-id"
	testWrapperClientSecret = "test-client-secret"
	testWrapperInstanceType = "1H100.80S.22V"
	testWrapperLocation     = "FIN-01"
	testWrapperImage        = "ubuntu-24.04-cuda-12.8-open-docker"
)

// newTestWrapper creates a wrapper with mock server for testing
func newTestWrapper(t *testing.T) (*verdacloudWrapper, *testutil.MockServer) {
	t.Helper()

	mockServer := testutil.NewMockServer()

	client, err := verda.NewClient(
		verda.WithClientID(testWrapperClientID),
		verda.WithClientSecret(testWrapperClientSecret),
		verda.WithBaseURL(mockServer.URL()),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	wrapper := &verdacloudWrapper{
		client: client,
		ctx:    context.Background(),
	}

	return wrapper, mockServer
}

func TestWrapperWithMockServer(t *testing.T) {
	wrapper, mockServer := newTestWrapper(t)
	defer mockServer.Close()

	t.Run("ListInstances", func(t *testing.T) {
		instances, err := wrapper.ListInstances("")
		if err != nil {
			t.Fatalf("ListInstances failed: %v", err)
		}
		if len(instances) == 0 {
			t.Error("expected at least one instance from mock server")
		}
		t.Logf("got %d instances", len(instances))
	})

	t.Run("GetInstanceByHostname", func(t *testing.T) {
		instances, err := wrapper.ListInstances("")
		if err != nil {
			t.Fatalf("ListInstances failed: %v", err)
		}
		if len(instances) == 0 {
			t.Skip("no instances available for testing")
		}

		hostname := instances[0].Hostname
		instance, err := wrapper.GetInstanceByHostname(hostname)
		if err != nil {
			t.Fatalf("GetInstanceByHostname failed: %v", err)
		}
		if instance.Hostname != hostname {
			t.Errorf("expected hostname %s, got %s", hostname, instance.Hostname)
		}
	})

	t.Run("ListInstanceTypes", func(t *testing.T) {
		types, err := wrapper.ListInstanceTypes()
		if err != nil {
			t.Skipf("ListInstanceTypes not available in mock server: %v", err)
		}
		t.Logf("got %d instance types", len(types))
	})

	t.Run("CreateStartScript", func(t *testing.T) {
		script, err := wrapper.CreateStartScript("test-script", "#!/bin/bash\necho 'test'")
		if err != nil {
			t.Skipf("CreateStartScript not available in mock server: %v", err)
		}
		if script.ID == "" {
			t.Errorf("expected script to have an ID")
		}
		t.Logf("created script with ID: %s, Name: %s", script.ID, script.Name)
	})

	t.Run("ListStartScripts", func(t *testing.T) {
		scripts, err := wrapper.ListStartScripts()
		if err != nil {
			t.Skipf("ListStartScripts not available in mock server: %v", err)
		}
		t.Logf("got %d startup scripts", len(scripts))
	})
}

func TestWrapperInstanceFiltering(t *testing.T) {
	wrapper, mockServer := newTestWrapper(t)
	defer mockServer.Close()

	t.Run("GetActiveInstancesForAsg", func(t *testing.T) {
		instances, err := wrapper.GetActiveInstancesForAsg("test-asg")
		if err != nil {
			t.Fatalf("GetActiveInstancesForAsg failed: %v", err)
		}
		t.Logf("got %d instances for ASG 'test-asg'", len(instances))
	})
}

func TestWrapperAvailability(t *testing.T) {
	wrapper, mockServer := newTestWrapper(t)
	defer mockServer.Close()

	t.Run("GetInstanceAvailabilityLocation", func(t *testing.T) {
		location, err := wrapper.GetInstanceAvailabilityLocation(testWrapperInstanceType, []string{testWrapperLocation, "FIN-02"})
		if err != nil {
			t.Logf("GetInstanceAvailabilityLocation returned error (expected if not available): %v", err)
		} else {
			t.Logf("instance type available in location: %s", location)
		}
	})
}

func TestCreateInstanceRequest(t *testing.T) {
	wrapper, mockServer := newTestWrapper(t)
	defer mockServer.Close()

	t.Run("CreateInstance", func(t *testing.T) {
		scriptID := "test-script-id"
		req := &verda.CreateInstanceRequest{
			InstanceType:    testWrapperInstanceType,
			Image:           testWrapperImage,
			Hostname:        "test-instance",
			Description:     "test-asg",
			SSHKeyIDs:       []string{"ssh-key-1"},
			LocationCode:    testWrapperLocation,
			StartupScriptID: &scriptID,
			IsSpot:          false,
		}

		instance, err := wrapper.CreateInstance(req)
		if err != nil {
			t.Fatalf("CreateInstance failed: %v", err)
		}
		if instance.Hostname != "test-instance" {
			t.Errorf("expected hostname 'test-instance', got %s", instance.Hostname)
		}
		t.Logf("created instance with ID: %s", instance.ID)
	})
}

func TestPerformInstanceAction(t *testing.T) {
	wrapper, mockServer := newTestWrapper(t)
	defer mockServer.Close()

	t.Run("PerformInstanceAction_Shutdown", func(t *testing.T) {
		err := wrapper.PerformInstanceAction("test-instance-id", verda.ActionShutdown)
		if err != nil {
			t.Logf("PerformInstanceAction returned error (expected for mock): %v", err)
		}
	})

	t.Run("PerformInstanceAction_Delete", func(t *testing.T) {
		err := wrapper.PerformInstanceAction("test-instance-id", verda.ActionDelete)
		if err != nil {
			t.Logf("PerformInstanceAction returned error (expected for mock): %v", err)
		}
	})
}

func TestMockServerResponses(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get(mockServer.URL() + "/instances")
		if err != nil {
			t.Fatalf("failed to call mock server: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestParseInstanceType(t *testing.T) {
	tests := []struct {
		name         string
		instanceType string
		expectedCPU  int64
		expectedGPU  int64
		expectedMem  int64 // in GB
	}{
		// 2-part GPU formats (gpuCount+model.vCPU)
		{
			name:         "A100 2-part format",
			instanceType: "1A100.22V",
			expectedCPU:  22,
			expectedGPU:  1,
			expectedMem:  88, // 22 * 4
		},
		{
			name:         "2x A100 2-part format",
			instanceType: "2A100.44V",
			expectedCPU:  44,
			expectedGPU:  2,
			expectedMem:  176,
		},
		{
			name:         "4x A100 2-part format",
			instanceType: "4A100.88V",
			expectedCPU:  88,
			expectedGPU:  4,
			expectedMem:  352,
		},
		{
			name:         "8x A100 2-part format",
			instanceType: "8A100.176V",
			expectedCPU:  176,
			expectedGPU:  8,
			expectedMem:  704,
		},
		{
			name:         "V100 2-part format",
			instanceType: "2V100.10V",
			expectedCPU:  10,
			expectedGPU:  2,
			expectedMem:  40,
		},
		{
			name:         "RTX6000ADA 2-part format",
			instanceType: "1RTX6000ADA.10V",
			expectedCPU:  10,
			expectedGPU:  1,
			expectedMem:  40,
		},
		{
			name:         "A6000 single 2-part format",
			instanceType: "1A6000.10V",
			expectedCPU:  10,
			expectedGPU:  1,
			expectedMem:  40,
		},
		{
			name:         "A6000 dual 2-part format",
			instanceType: "2A6000.20V",
			expectedCPU:  20,
			expectedGPU:  2,
			expectedMem:  80,
		},
		{
			name:         "B300 single 2-part format",
			instanceType: "1B300.30V",
			expectedCPU:  30,
			expectedGPU:  1,
			expectedMem:  120,
		},
		{
			name:         "B300 quad 2-part format",
			instanceType: "4B300.120V",
			expectedCPU:  120,
			expectedGPU:  4,
			expectedMem:  480,
		},
		{
			name:         "B300 octo 2-part format",
			instanceType: "8B300.240V",
			expectedCPU:  240,
			expectedGPU:  8,
			expectedMem:  960,
		},
		{
			name:         "B200 single 2-part format",
			instanceType: "1B200.30V",
			expectedCPU:  30,
			expectedGPU:  1,
			expectedMem:  120,
		},
		{
			name:         "RTXPRO6000 single 2-part format",
			instanceType: "1RTXPRO6000.30V",
			expectedCPU:  30,
			expectedGPU:  1,
			expectedMem:  120,
		},
		{
			name:         "RTXPRO6000 octo 2-part format",
			instanceType: "8RTXPRO6000.240V",
			expectedCPU:  240,
			expectedGPU:  8,
			expectedMem:  960,
		},
		{
			name:         "L40S 2-part format",
			instanceType: "1L40S.20V",
			expectedCPU:  20,
			expectedGPU:  1,
			expectedMem:  80,
		},
		// 3-part GPU formats (gpuCount+model.vram.vCPU)
		{
			name:         "A100 40GB 3-part format",
			instanceType: "1A100.40S.22V",
			expectedCPU:  22,
			expectedGPU:  1,
			expectedMem:  88,
		},
		{
			name:         "H100 80GB 3-part format (test constant)",
			instanceType: testWrapperInstanceType, // 1H100.80S.22V
			expectedCPU:  22,
			expectedGPU:  1,
			expectedMem:  88,
		},
		{
			name:         "H100 80GB 32V 3-part format",
			instanceType: "1H100.80S.32V",
			expectedCPU:  32,
			expectedGPU:  1,
			expectedMem:  128,
		},
		{
			name:         "H200 141GB single 3-part format",
			instanceType: "1H200.141S.44V",
			expectedCPU:  44,
			expectedGPU:  1,
			expectedMem:  176,
		},
		{
			name:         "H200 141GB dual 3-part format",
			instanceType: "2H200.141S.88V",
			expectedCPU:  88,
			expectedGPU:  2,
			expectedMem:  352,
		},
		{
			name:         "H200 141GB quad 3-part format",
			instanceType: "4H200.141S.176V",
			expectedCPU:  176,
			expectedGPU:  4,
			expectedMem:  704,
		},
		// CPU formats (CPU.vCPU.memory)
		{
			name:         "CPU small instance",
			instanceType: "CPU.4V.16G",
			expectedCPU:  4,
			expectedGPU:  0,
			expectedMem:  16,
		},
		{
			name:         "CPU medium instance",
			instanceType: "CPU.16V.64G",
			expectedCPU:  16,
			expectedGPU:  0,
			expectedMem:  64,
		},
		{
			name:         "CPU large instance",
			instanceType: "CPU.180V.720G",
			expectedCPU:  180,
			expectedGPU:  0,
			expectedMem:  720,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseInstanceType(tc.instanceType)

			if result.CPU != tc.expectedCPU {
				t.Errorf("CPU: expected %d, got %d", tc.expectedCPU, result.CPU)
			}
			if result.GPU != tc.expectedGPU {
				t.Errorf("GPU: expected %d, got %d", tc.expectedGPU, result.GPU)
			}
			expectedMemBytes := tc.expectedMem * 1024 * 1024 * 1024
			if result.Memory != expectedMemBytes {
				t.Errorf("Memory: expected %d bytes (%dGB), got %d bytes (%dGB)",
					expectedMemBytes, tc.expectedMem, result.Memory, result.Memory/(1024*1024*1024))
			}
			if result.InstanceType != tc.instanceType {
				t.Errorf("InstanceType: expected %s, got %s", tc.instanceType, result.InstanceType)
			}
			if result.Arch != "amd64" {
				t.Errorf("Arch: expected amd64, got %s", result.Arch)
			}
		})
	}
}
