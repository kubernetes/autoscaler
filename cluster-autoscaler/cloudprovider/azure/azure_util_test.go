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
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"k8s.io/legacy-cloud-providers/azure/clients/diskclient/mockdiskclient"
	"k8s.io/legacy-cloud-providers/azure/clients/interfaceclient/mockinterfaceclient"
	"k8s.io/legacy-cloud-providers/azure/clients/storageaccountclient/mockstorageaccountclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmclient/mockvmclient"
	"k8s.io/legacy-cloud-providers/azure/retry"
)

const (
	testAccountName            = "account"
	storageAccountClientErrMsg = "Server failed to authenticate the request. Make sure the value of Authorization " +
		"header is formed correctly including the signature"
)

func GetTestAzureUtil(t *testing.T) *AzUtil {
	return &AzUtil{manager: newTestAzureManager(t)}
}

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
			t.Fatalf("incorrect aks string. expected=%s actual=%s", d.expectedOrch, orch)
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

func TestIsSuccessResponse(t *testing.T) {
	tests := []struct {
		name          string
		resp          *http.Response
		err           error
		expected      bool
		expectedError error
	}{
		{
			name:          "both resp and err nil should report error",
			expected:      false,
			expectedError: fmt.Errorf("failed with unknown error"),
		},
		{
			name: "http.StatusNotFound should report error",
			resp: &http.Response{
				StatusCode: http.StatusNotFound,
			},
			expected:      false,
			expectedError: fmt.Errorf("failed with HTTP status code %d", http.StatusNotFound),
		},
		{
			name: "http.StatusInternalServerError should report error",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			expected:      false,
			expectedError: fmt.Errorf("failed with HTTP status code %d", http.StatusInternalServerError),
		},
		{
			name: "http.StatusOK shouldn't report error",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			expected: true,
		},
		{
			name: "non-nil response error with http.StatusOK should report error",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			err:           fmt.Errorf("test error"),
			expected:      false,
			expectedError: fmt.Errorf("test error"),
		},
		{
			name: "non-nil response error with http.StatusInternalServerError should report error",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			err:           fmt.Errorf("test error"),
			expected:      false,
			expectedError: fmt.Errorf("test error"),
		},
	}

	for _, test := range tests {
		result, realError := isSuccessHTTPResponse(test.resp, test.err)
		assert.Equal(t, test.expected, result, "[%s] expected: %v, saw: %v", test.name, result, test.expected)
		assert.Equal(t, test.expectedError, realError, "[%s] expected: %v, saw: %v", test.name, realError, test.expectedError)
	}
}
func TestConvertResourceGroupNameToLower(t *testing.T) {
	tests := []struct {
		desc        string
		resourceID  string
		expected    string
		expectError bool
	}{
		{
			desc:        "empty string should report error",
			resourceID:  "",
			expectError: true,
		},
		{
			desc:        "resourceID not in Azure format should report error",
			resourceID:  "invalid-id",
			expectError: true,
		},
		{
			desc:        "providerID not in Azure format should report error",
			resourceID:  "azure://invalid-id",
			expectError: true,
		},
		{
			desc:       "resource group name in VM providerID should be converted",
			resourceID: "azure:///subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroupName/providers/Microsoft.Compute/virtualMachines/k8s-agent-AAAAAAAA-0",
			expected:   "azure:///subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroupname/providers/Microsoft.Compute/virtualMachines/k8s-agent-AAAAAAAA-0",
		},
		{
			desc:       "resource group name in VM resourceID should be converted",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroupName/providers/Microsoft.Compute/virtualMachines/k8s-agent-AAAAAAAA-0",
			expected:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroupname/providers/Microsoft.Compute/virtualMachines/k8s-agent-AAAAAAAA-0",
		},
		{
			desc:       "resource group name in VMSS providerID should be converted",
			resourceID: "azure:///subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroupName/providers/Microsoft.Compute/virtualMachineScaleSets/myScaleSetName/virtualMachines/156",
			expected:   "azure:///subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroupname/providers/Microsoft.Compute/virtualMachineScaleSets/myScaleSetName/virtualMachines/156",
		},
		{
			desc:       "resource group name in VMSS resourceID should be converted",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroupName/providers/Microsoft.Compute/virtualMachineScaleSets/myScaleSetName/virtualMachines/156",
			expected:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroupname/providers/Microsoft.Compute/virtualMachineScaleSets/myScaleSetName/virtualMachines/156",
		},
	}

	for _, test := range tests {
		real, err := convertResourceGroupNameToLower(test.resourceID)
		if test.expectError {
			assert.NotNil(t, err, test.desc)
			continue
		}

		assert.Nil(t, err, test.desc)
		assert.Equal(t, test.expected, real, test.desc)
	}
}

func TestIsAzureRequestsThrottled(t *testing.T) {
	tests := []struct {
		desc     string
		rerr     *retry.Error
		expected bool
	}{
		{
			desc:     "nil error should return false",
			expected: false,
		},
		{
			desc: "non http.StatusTooManyRequests error should return false",
			rerr: &retry.Error{
				HTTPStatusCode: http.StatusBadRequest,
			},
			expected: false,
		},
		{
			desc: "http.StatusTooManyRequests error should return true",
			rerr: &retry.Error{
				HTTPStatusCode: http.StatusTooManyRequests,
			},
			expected: true,
		},
		{
			desc: "Nul HTTP code and non-expired Retry-After should return true",
			rerr: &retry.Error{
				RetryAfter: time.Now().Add(time.Hour),
			},
			expected: true,
		},
	}

	for _, test := range tests {
		real := isAzureRequestsThrottled(test.rerr)
		assert.Equal(t, test.expected, real, test.desc)
	}
}

func TestDeleteBlob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	azUtil := GetTestAzureUtil(t)
	mockSAClient := mockstorageaccountclient.NewMockInterface(ctrl)
	mockSAClient.EXPECT().ListKeys(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		testAccountName).Return(storage.AccountListKeysResult{
		Keys: &[]storage.AccountKey{
			{Value: to.StringPtr("dmFsdWUK")},
		},
	}, nil)
	azUtil.manager.azClient.storageAccountsClient = mockSAClient

	err := azUtil.DeleteBlob(testAccountName, "vhd", "blob")
	assert.True(t, strings.Contains(err.Error(), storageAccountClientErrMsg))
}

func TestDeleteVirtualMachine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	azUtil := GetTestAzureUtil(t)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	azUtil.manager.azClient.virtualMachinesClient = mockVMClient

	mockVMClient.EXPECT().Get(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"vm",
		gomock.Any()).Return(compute.VirtualMachine{}, errInternal)

	err := azUtil.DeleteVirtualMachine("rg", "vm")
	assert.NoError(t, err)

	mockVMClient.EXPECT().Get(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"vm",
		gomock.Any()).Return(compute.VirtualMachine{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{},
			},
		},
	}, nil)
	err = azUtil.DeleteVirtualMachine("rg", "vm")
	expectedErr := fmt.Errorf("os disk does not have a VHD URI")
	assert.Equal(t, expectedErr, err)

	mockVMClient.EXPECT().Get(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"vm",
		gomock.Any()).Return(compute.VirtualMachine{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					Vhd: &compute.VirtualHardDisk{
						URI: to.StringPtr("https://vhdstorage8h8pjybi9hbsl6.blob.core.windows.net" +
							"/vhds/osdisks/disk1234.vhd"),
					},
				},
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{ID: to.StringPtr("foo/bar")},
				},
			},
		},
	}, nil)
	mockVMClient.EXPECT().Delete(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"vm").Return(nil).Times(2)
	mockSAClient := mockstorageaccountclient.NewMockInterface(ctrl)
	mockSAClient.EXPECT().ListKeys(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"vhdstorage8h8pjybi9hbsl6").Return(storage.AccountListKeysResult{
		Keys: &[]storage.AccountKey{
			{Value: to.StringPtr("dmFsdWUK")},
		},
	}, nil)
	azUtil.manager.azClient.storageAccountsClient = mockSAClient
	mockNICClient := mockinterfaceclient.NewMockInterface(ctrl)
	mockNICClient.EXPECT().Delete(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"bar").Return(nil).Times(2)
	azUtil.manager.azClient.interfacesClient = mockNICClient
	err = azUtil.DeleteVirtualMachine("rg", "vm")
	assert.True(t, strings.Contains(err.Error(), "no such host"))

	mockVMClient.EXPECT().Get(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"vm",
		gomock.Any()).Return(compute.VirtualMachine{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					Name:        to.StringPtr("disk"),
					ManagedDisk: &compute.ManagedDiskParameters{},
				},
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{ID: to.StringPtr("foo/bar")},
				},
			},
		},
	}, nil)
	mockDiskClient := mockdiskclient.NewMockInterface(ctrl)
	mockDiskClient.EXPECT().Delete(
		gomock.Any(),
		azUtil.manager.config.ResourceGroup,
		"disk").Return(nil)
	azUtil.manager.azClient.disksClient = mockDiskClient
	err = azUtil.DeleteVirtualMachine("rg", "vm")
	assert.NoError(t, err)
}

func TestNormalizeMasterResourcesForScaling(t *testing.T) {
	templateMap := map[string]interface{}{
		resourcesFieldName: []interface{}{
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmExtensionType,
			},
			map[string]interface{}{
				nameFieldName: 1,
				typeFieldName: vmResourceType,
			},
			map[string]interface{}{
				nameFieldName: "foo",
				typeFieldName: vmResourceType,
			},
			map[string]interface{}{
				nameFieldName:       "variables('masterVMNamePrefix')",
				typeFieldName:       vmResourceType,
				propertiesFieldName: "foo",
			},
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmResourceType,
				propertiesFieldName: map[string]interface{}{
					hardwareProfileFieldName: "foo",
				},
			},
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmResourceType,
				propertiesFieldName: map[string]interface{}{
					hardwareProfileFieldName: map[string]interface{}{
						vmSizeFieldName: "size",
					},
				},
			},
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmResourceType,
				propertiesFieldName: map[string]interface{}{
					hardwareProfileFieldName: map[string]interface{}{},
					osProfileFieldName:       "foo",
				},
			},
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmResourceType,
				propertiesFieldName: map[string]interface{}{
					hardwareProfileFieldName: map[string]interface{}{},
					osProfileFieldName: map[string]interface{}{
						customDataFieldName: "data",
					},
				},
			},
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmResourceType,
				propertiesFieldName: map[string]interface{}{
					hardwareProfileFieldName: map[string]interface{}{},
					storageProfileFieldName:  "foo",
				},
			},
			map[string]interface{}{
				nameFieldName: "variables('masterVMNamePrefix')",
				typeFieldName: vmResourceType,
				propertiesFieldName: map[string]interface{}{
					hardwareProfileFieldName: map[string]interface{}{},
					storageProfileFieldName: map[string]interface{}{
						imageReferenceFieldName: "image",
					},
				},
			},
		},
	}
	err := normalizeMasterResourcesForScaling(templateMap)
	assert.Equal(t, 9, len(templateMap[resourcesFieldName].([]interface{})))
	assert.NoError(t, err)
}

func TestNormalizeForK8sVMASScalingUp(t *testing.T) {
	templateMap := map[string]interface{}{
		resourcesFieldName: []interface{}{
			map[string]interface{}{
				typeFieldName: nsgResourceType,
			},
			map[string]interface{}{
				typeFieldName: nsgResourceType,
			},
		},
	}
	err := normalizeForK8sVMASScalingUp(templateMap)
	expectedErr := fmt.Errorf("found 2 resources with type %s in the template. "+
		"There should only be 1", nsgResourceType)
	assert.Equal(t, expectedErr, err)

	templateMap = map[string]interface{}{
		resourcesFieldName: []interface{}{
			map[string]interface{}{
				typeFieldName: rtResourceType,
			},
			map[string]interface{}{
				typeFieldName: rtResourceType,
			},
		},
	}
	expectedErr = fmt.Errorf("found 2 resources with type %s in the template. "+
		"There should only be 1", rtResourceType)
	err = normalizeForK8sVMASScalingUp(templateMap)
	assert.Equal(t, expectedErr, err)

	templateMap = map[string]interface{}{
		resourcesFieldName: []interface{}{
			map[string]interface{}{
				typeFieldName: nsgResourceType,
			},
			map[string]interface{}{
				dependsOnFieldName: []interface{}{nsgResourceType, "foo"},
			},
		},
	}
	err = normalizeForK8sVMASScalingUp(templateMap)
	for _, resource := range templateMap[resourcesFieldName].([]interface{}) {
		deps, ok := resource.([]interface{})
		if ok {
			for _, dep := range deps {
				if names, ok := dep.(map[string]interface{})[dependsOnFieldName]; ok {
					assert.Equal(t, 1, len(names.([]interface{})))
				}
			}
		}
	}
	assert.NoError(t, err)
}
