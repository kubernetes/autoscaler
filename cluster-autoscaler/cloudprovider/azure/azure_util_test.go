/*
Copyright 2017 The Kubernetes Authors.

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

package azure

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
)

func TestSplitBlobURI(t *testing.T) {
	expectedAccountName := "vhdstorage8h8pjybi9hbsl6"
	expectedContainerName := "vhds"
	expectedBlobPath := "osdisks/disk1234.vhd"
	accountName, containerName, blobPath, err := splitBlobURI("https://vhdstorage8h8pjybi9hbsl6.blob.core.windows.net/vhds/osdisks/disk1234.vhd")
	if accountName != expectedAccountName {
		t.Fatalf("incorrect account name. expected=%s actual=%s", expectedAccountName, accountName)
	}
	if containerName != expectedContainerName {
		t.Fatalf("incorrect account name. expected=%s actual=%s", expectedContainerName, containerName)
	}
	if blobPath != expectedBlobPath {
		t.Fatalf("incorrect account name. expected=%s actual=%s", expectedBlobPath, blobPath)
	}
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestK8sLinuxVMNameParts(t *testing.T) {
	data := []struct {
		poolIdentifier, nameSuffix string
		agentIndex                 int
	}{
		{"agentpool1", "38988164", 10},
		{"agent-pool1", "38988164", 8},
		{"agent-pool-1", "38988164", 0},
	}

	for _, el := range data {
		vmName := fmt.Sprintf("k8s-%s-%s-%d", el.poolIdentifier, el.nameSuffix, el.agentIndex)
		poolIdentifier, nameSuffix, agentIndex, err := k8sLinuxVMNameParts(vmName)
		if poolIdentifier != el.poolIdentifier {
			t.Fatalf("incorrect poolIdentifier. expected=%s actual=%s", el.poolIdentifier, poolIdentifier)
		}
		if nameSuffix != el.nameSuffix {
			t.Fatalf("incorrect nameSuffix. expected=%s actual=%s", el.nameSuffix, nameSuffix)
		}
		if agentIndex != el.agentIndex {
			t.Fatalf("incorrect agentIndex. expected=%d actual=%d", el.agentIndex, agentIndex)
		}
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}
}

func TestWindowsVMNameParts(t *testing.T) {
	data := []struct {
		VMName, expectedPoolPrefix, expectedOrch string
		expectedPoolIndex, expectedAgentIndex    int
	}{
		{"38988k8s90312", "38988", "k8s", 3, 12},
		{"4506k8s010", "4506", "k8s", 1, 0},
		{"2314k8s03000001", "2314", "k8s", 3, 1},
		{"2314k8s0310", "2314", "k8s", 3, 10},
	}

	for _, d := range data {
		poolPrefix, orch, poolIndex, agentIndex, err := windowsVMNameParts(d.VMName)
		if poolPrefix != d.expectedPoolPrefix {
			t.Fatalf("incorrect poolPrefix. expected=%s actual=%s", d.expectedPoolPrefix, poolPrefix)
		}
		if orch != d.expectedOrch {
			t.Fatalf("incorrect acs string. expected=%s actual=%s", d.expectedOrch, orch)
		}
		if poolIndex != d.expectedPoolIndex {
			t.Fatalf("incorrect poolIndex. expected=%d actual=%d", d.expectedPoolIndex, poolIndex)
		}
		if agentIndex != d.expectedAgentIndex {
			t.Fatalf("incorrect agentIndex. expected=%d actual=%d", d.expectedAgentIndex, agentIndex)
		}
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}
}

func TestGetVMNameIndexLinux(t *testing.T) {
	expectedAgentIndex := 65

	agentIndex, err := GetVMNameIndex(compute.Linux, "k8s-agentpool1-38988164-65")
	if agentIndex != expectedAgentIndex {
		t.Fatalf("incorrect agentIndex. expected=%d actual=%d", expectedAgentIndex, agentIndex)
	}
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestGetVMNameIndexWindows(t *testing.T) {
	expectedAgentIndex := 20

	agentIndex, err := GetVMNameIndex(compute.Windows, "38988k8s90320")
	if agentIndex != expectedAgentIndex {
		t.Fatalf("incorrect agentIndex. expected=%d actual=%d", expectedAgentIndex, agentIndex)
	}
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
