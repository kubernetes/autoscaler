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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	azStorage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	klog "k8s.io/klog/v2"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	//Field names
	customDataFieldName      = "customData"
	dependsOnFieldName       = "dependsOn"
	hardwareProfileFieldName = "hardwareProfile"
	imageReferenceFieldName  = "imageReference"
	nameFieldName            = "name"
	osProfileFieldName       = "osProfile"
	propertiesFieldName      = "properties"
	resourcesFieldName       = "resources"
	storageProfileFieldName  = "storageProfile"
	typeFieldName            = "type"
	vmSizeFieldName          = "vmSize"

	// ARM resource Types
	nsgResourceType = "Microsoft.Network/networkSecurityGroups"
	rtResourceType  = "Microsoft.Network/routeTables"
	vmResourceType  = "Microsoft.Compute/virtualMachines"
	vmExtensionType = "Microsoft.Compute/virtualMachines/extensions"

	// CSE Extension checks
	vmssCSEExtensionName            = "vmssCSE"
	vmssExtensionProvisioningFailed = "VMExtensionProvisioningFailed"
	// vmExtensionProvisioningErrorClass represents a Vm extension provisioning error
	vmExtensionProvisioningErrorClass cloudprovider.InstanceErrorClass = 103

	// resource ids
	nsgID = "nsgID"
	rtID  = "routeTableID"

	k8sLinuxVMNamingFormat         = "^[0-9a-zA-Z]{3}-(.+)-([0-9a-fA-F]{8})-{0,2}([0-9]+)$"
	k8sLinuxVMAgentPoolNameIndex   = 1
	k8sLinuxVMAgentClusterIDIndex  = 2
	k8sLinuxVMAgentIndexArrayIndex = 3

	k8sWindowsOldVMNamingFormat            = "^([a-fA-F0-9]{5})([0-9a-zA-Z]{3})([9])([a-zA-Z0-9]{3,5})$"
	k8sWindowsVMNamingFormat               = "^([a-fA-F0-9]{4})([0-9a-zA-Z]{3})([0-9]{3,8})$"
	k8sWindowsVMAgentPoolPrefixIndex       = 1
	k8sWindowsVMAgentOrchestratorNameIndex = 2
	k8sWindowsVMAgentPoolInfoIndex         = 3

	nodeLabelTagName     = "k8s.io_cluster-autoscaler_node-template_label_"
	nodeTaintTagName     = "k8s.io_cluster-autoscaler_node-template_taint_"
	nodeResourcesTagName = "k8s.io_cluster-autoscaler_node-template_resources_"
	nodeOptionsTagName   = "k8s.io_cluster-autoscaler_node-template_autoscaling-options_"

	// PowerStates reflect the operational state of a VM
	// From https://learn.microsoft.com/en-us/java/api/com.microsoft.azure.management.compute.powerstate?view=azure-java-stable
	vmPowerStateStarting     = "PowerState/starting"
	vmPowerStateRunning      = "PowerState/running"
	vmPowerStateStopping     = "PowerState/stopping"
	vmPowerStateStopped      = "PowerState/stopped"
	vmPowerStateDeallocating = "PowerState/deallocating"
	vmPowerStateDeallocated  = "PowerState/deallocated"
	vmPowerStateUnknown      = "PowerState/unknown"
)

var (
	vmnameLinuxRegexp        = regexp.MustCompile(k8sLinuxVMNamingFormat)
	vmnameWindowsRegexp      = regexp.MustCompile(k8sWindowsVMNamingFormat)
	oldvmnameWindowsRegexp   = regexp.MustCompile(k8sWindowsOldVMNamingFormat)
	azureResourceGroupNameRE = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/(?:.*)`)
)

// AzUtil consists of utility functions which utilizes clients to different services.
// Since they span across various clients they cannot be fitted into individual client structs
// so adding them here.
type AzUtil struct {
	manager *AzureManager
}

// DeleteBlob deletes the blob using the storage client.
func (util *AzUtil) DeleteBlob(accountName, vhdContainer, vhdBlob string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	storageKeysResult, rerr := util.manager.azClient.storageAccountsClient.ListKeys(ctx, util.manager.config.SubscriptionID, util.manager.config.ResourceGroup, accountName)
	if rerr != nil {
		return rerr.Error()
	}

	keys := *storageKeysResult.Keys
	client, err := azStorage.NewBasicClientOnSovereignCloud(accountName, to.String(keys[0].Value), util.manager.env)
	if err != nil {
		return err
	}

	bs := client.GetBlobService()
	containerRef := bs.GetContainerReference(vhdContainer)
	blobRef := containerRef.GetBlobReference(vhdBlob)

	return blobRef.Delete(&azStorage.DeleteBlobOptions{})
}

// DeleteVirtualMachine deletes a VM and any associated OS disk
func (util *AzUtil) DeleteVirtualMachine(rg string, name string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	vm, rerr := util.manager.azClient.virtualMachinesClient.Get(ctx, rg, name, "")
	if rerr != nil {
		if exists, _ := checkResourceExistsFromRetryError(rerr); !exists {
			klog.V(2).Infof("VirtualMachine %s/%s has already been removed", rg, name)
			return nil
		}

		klog.Errorf("failed to get VM: %s/%s: %s", rg, name, rerr.Error())
		return rerr.Error()
	}

	vhd := vm.VirtualMachineProperties.StorageProfile.OsDisk.Vhd
	managedDisk := vm.VirtualMachineProperties.StorageProfile.OsDisk.ManagedDisk
	if vhd == nil && managedDisk == nil {
		klog.Errorf("failed to get a valid os disk URI for VM: %s/%s", rg, name)
		return fmt.Errorf("os disk does not have a VHD URI")
	}

	osDiskName := vm.VirtualMachineProperties.StorageProfile.OsDisk.Name
	var nicName string
	var err error
	nicID := (*vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces)[0].ID
	if nicID == nil {
		klog.Warningf("NIC ID is not set for VM (%s/%s)", rg, name)
	} else {
		nicName, err = resourceName(*nicID)
		if err != nil {
			return err
		}
		klog.Infof("found nic name for VM (%s/%s): %s", rg, name, nicName)
	}

	klog.Infof("deleting VM: %s/%s", rg, name)
	deleteCtx, deleteCancel := getContextWithCancel()
	defer deleteCancel()

	klog.Infof("waiting for VirtualMachine deletion: %s/%s", rg, name)
	rerr = util.manager.azClient.virtualMachinesClient.Delete(deleteCtx, rg, name)
	_, realErr := checkResourceExistsFromRetryError(rerr)
	if realErr != nil {
		return realErr
	}
	klog.V(2).Infof("VirtualMachine %s/%s removed", rg, name)

	if nicName != "" {
		klog.Infof("deleting nic: %s/%s", rg, nicName)
		interfaceCtx, interfaceCancel := getContextWithCancel()
		defer interfaceCancel()
		klog.Infof("waiting for nic deletion: %s/%s", rg, nicName)
		nicErr := util.manager.azClient.interfacesClient.Delete(interfaceCtx, rg, nicName)
		_, realErr := checkResourceExistsFromRetryError(nicErr)
		if realErr != nil {
			return realErr
		}
		klog.V(2).Infof("interface %s/%s removed", rg, nicName)
	}

	if vhd != nil {
		accountName, vhdContainer, vhdBlob, err := splitBlobURI(*vhd.URI)
		if err != nil {
			return err
		}

		klog.Infof("found os disk storage reference: %s %s %s", accountName, vhdContainer, vhdBlob)

		klog.Infof("deleting blob: %s/%s", vhdContainer, vhdBlob)
		if err = util.DeleteBlob(accountName, vhdContainer, vhdBlob); err != nil {
			_, realErr := checkResourceExistsFromError(err)
			if realErr != nil {
				return realErr
			}
			klog.V(2).Infof("Blob %s/%s removed", rg, vhdBlob)
		}
	} else if managedDisk != nil {
		if osDiskName == nil {
			klog.Warningf("osDisk is not set for VM %s/%s", rg, name)
		} else {
			klog.Infof("deleting managed disk: %s/%s", rg, *osDiskName)
			disksCtx, disksCancel := getContextWithCancel()
			defer disksCancel()
			diskErr := util.manager.azClient.disksClient.Delete(disksCtx, util.manager.config.SubscriptionID, rg, *osDiskName)
			_, realErr := checkResourceExistsFromRetryError(diskErr)
			if realErr != nil {
				return realErr
			}
			klog.V(2).Infof("disk %s/%s removed", rg, *osDiskName)
		}
	}
	return nil
}

func getUserAgentExtension() string {
	suffix := os.Getenv("AZURE_CLUSTER_AUTOSCALER_USER_AGENT_SUFFIX")
	return fmt.Sprintf("cluster-autoscaler%s/v%s", suffix, version.ClusterAutoscalerVersion)
}

func configureUserAgent(client *autorest.Client) {
	client.UserAgent = fmt.Sprintf("%s; %s", client.UserAgent, getUserAgentExtension())
}

// normalizeForK8sVMASScalingUp takes a template and removes elements that are unwanted in a K8s VMAS scale up/down case
func normalizeForK8sVMASScalingUp(templateMap map[string]interface{}) error {
	if err := normalizeMasterResourcesForScaling(templateMap); err != nil {
		return err
	}
	rtIndex := -1
	nsgIndex := -1
	resources := templateMap[resourcesFieldName].([]interface{})
	for index, resource := range resources {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			klog.Warning("Template improperly formatted for resource")
			continue
		}

		resourceType, ok := resourceMap[typeFieldName].(string)
		if ok && resourceType == nsgResourceType {
			if nsgIndex != -1 {
				err := fmt.Errorf("found 2 resources with type %s in the template. There should only be 1", nsgResourceType)
				klog.Errorf(err.Error())
				return err
			}
			nsgIndex = index
		}
		if ok && resourceType == rtResourceType {
			if rtIndex != -1 {
				err := fmt.Errorf("found 2 resources with type %s in the template. There should only be 1", rtResourceType)
				klog.Warningf(err.Error())
				return err
			}
			rtIndex = index
		}

		dependencies, ok := resourceMap[dependsOnFieldName].([]interface{})
		if !ok {
			continue
		}

		for dIndex := len(dependencies) - 1; dIndex >= 0; dIndex-- {
			dependency := dependencies[dIndex].(string)
			if strings.Contains(dependency, nsgResourceType) || strings.Contains(dependency, nsgID) ||
				strings.Contains(dependency, rtResourceType) || strings.Contains(dependency, rtID) {
				dependencies = append(dependencies[:dIndex], dependencies[dIndex+1:]...)
			}
		}

		if len(dependencies) > 0 {
			resourceMap[dependsOnFieldName] = dependencies
		} else {
			delete(resourceMap, dependsOnFieldName)
		}
	}

	indexesToRemove := []int{}
	if nsgIndex == -1 {
		err := fmt.Errorf("found no resources with type %s in the template. There should have been 1", nsgResourceType)
		klog.Errorf(err.Error())
		return err
	}
	if rtIndex == -1 {
		klog.Infof("Found no resources with type %s in the template.", rtResourceType)
	} else {
		indexesToRemove = append(indexesToRemove, rtIndex)
	}
	indexesToRemove = append(indexesToRemove, nsgIndex)
	templateMap[resourcesFieldName] = removeIndexesFromArray(resources, indexesToRemove)

	return nil
}

func removeIndexesFromArray(array []interface{}, indexes []int) []interface{} {
	sort.Sort(sort.Reverse(sort.IntSlice(indexes)))
	for _, index := range indexes {
		array = append(array[:index], array[index+1:]...)
	}
	return array
}

// normalizeMasterResourcesForScaling takes a template and removes elements that are unwanted in any scale up/down case
func normalizeMasterResourcesForScaling(templateMap map[string]interface{}) error {
	resources := templateMap[resourcesFieldName].([]interface{})
	indexesToRemove := []int{}
	//update master nodes resources
	for index, resource := range resources {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			klog.Warning("Template improperly formatted")
			continue
		}

		resourceType, ok := resourceMap[typeFieldName].(string)
		if !ok || resourceType != vmResourceType {
			resourceName, ok := resourceMap[nameFieldName].(string)
			if !ok {
				klog.Warning("Template improperly formatted")
				continue
			}
			if strings.Contains(resourceName, "variables('masterVMNamePrefix')") && resourceType == vmExtensionType {
				indexesToRemove = append(indexesToRemove, index)
			}
			continue
		}

		resourceName, ok := resourceMap[nameFieldName].(string)
		if !ok {
			klog.Warning("Template improperly formatted")
			continue
		}

		// make sure this is only modifying the master vms
		if !strings.Contains(resourceName, "variables('masterVMNamePrefix')") {
			continue
		}

		resourceProperties, ok := resourceMap[propertiesFieldName].(map[string]interface{})
		if !ok {
			klog.Warning("Template improperly formatted")
			continue
		}

		hardwareProfile, ok := resourceProperties[hardwareProfileFieldName].(map[string]interface{})
		if !ok {
			klog.Warning("Template improperly formatted")
			continue
		}

		if hardwareProfile[vmSizeFieldName] != nil {
			delete(hardwareProfile, vmSizeFieldName)
		}

		if !removeCustomData(resourceProperties) || !removeImageReference(resourceProperties) {
			continue
		}
	}
	templateMap[resourcesFieldName] = removeIndexesFromArray(resources, indexesToRemove)

	return nil
}

func removeCustomData(resourceProperties map[string]interface{}) bool {
	osProfile, ok := resourceProperties[osProfileFieldName].(map[string]interface{})
	if !ok {
		klog.Warning("Template improperly formatted")
		return ok
	}

	if osProfile[customDataFieldName] != nil {
		delete(osProfile, customDataFieldName)
	}
	return ok
}

func removeImageReference(resourceProperties map[string]interface{}) bool {
	storageProfile, ok := resourceProperties[storageProfileFieldName].(map[string]interface{})
	if !ok {
		klog.Warningf("Template improperly formatted. Could not find: %s", storageProfileFieldName)
		return ok
	}

	if storageProfile[imageReferenceFieldName] != nil {
		delete(storageProfile, imageReferenceFieldName)
	}
	return ok
}

// resourceName returns the last segment (the resource name) for the specified resource identifier.
func resourceName(ID string) (string, error) {
	parts := strings.Split(ID, "/")
	name := parts[len(parts)-1]
	if len(name) == 0 {
		return "", fmt.Errorf("resource name was missing from identifier")
	}

	return name, nil
}

// splitBlobURI returns a decomposed blob URI parts: accountName, containerName, blobName.
func splitBlobURI(URI string) (string, string, string, error) {
	uri, err := url.Parse(URI)
	if err != nil {
		return "", "", "", err
	}

	accountName := strings.Split(uri.Host, ".")[0]
	urlParts := strings.Split(uri.Path, "/")

	containerName := urlParts[1]
	blobPath := strings.Join(urlParts[2:], "/")

	return accountName, containerName, blobPath, nil
}

// k8sLinuxVMNameParts returns parts of Linux VM name e.g: k8s-agentpool1-11290731-0
func k8sLinuxVMNameParts(vmName string) (poolIdentifier, nameSuffix string, agentIndex int, err error) {
	vmNameParts := vmnameLinuxRegexp.FindStringSubmatch(vmName)
	if len(vmNameParts) != 4 {
		return "", "", -1, fmt.Errorf("resource name was missing from identifier")
	}

	vmNum, err := strconv.Atoi(vmNameParts[k8sLinuxVMAgentIndexArrayIndex])

	if err != nil {
		return "", "", -1, fmt.Errorf("error parsing VM Name: %v", err)
	}

	return vmNameParts[k8sLinuxVMAgentPoolNameIndex], vmNameParts[k8sLinuxVMAgentClusterIDIndex], vmNum, nil
}

// windowsVMNameParts returns parts of Windows VM name
func windowsVMNameParts(vmName string) (poolPrefix string, orch string, poolIndex int, agentIndex int, err error) {
	var poolInfo string
	vmNameParts := oldvmnameWindowsRegexp.FindStringSubmatch(vmName)
	if len(vmNameParts) != 5 {
		vmNameParts = vmnameWindowsRegexp.FindStringSubmatch(vmName)
		if len(vmNameParts) != 4 {
			return "", "", -1, -1, fmt.Errorf("resource name was missing from identifier")
		}
		poolInfo = vmNameParts[3]
	} else {
		poolInfo = vmNameParts[4]
	}

	poolPrefix = vmNameParts[1]
	orch = vmNameParts[2]

	poolIndex, err = strconv.Atoi(poolInfo[:2])
	if err != nil {
		return "", "", -1, -1, fmt.Errorf("error parsing VM Name: %v", err)
	}
	agentIndex, err = strconv.Atoi(poolInfo[2:])
	if err != nil {
		return "", "", -1, -1, fmt.Errorf("error parsing VM Name: %v", err)
	}

	return poolPrefix, orch, poolIndex, agentIndex, nil
}

// GetVMNameIndex return the index of VM in the node pools.
func GetVMNameIndex(osType compute.OperatingSystemTypes, vmName string) (int, error) {
	var agentIndex int
	var err error
	if osType == compute.OperatingSystemTypesLinux {
		_, _, agentIndex, err = k8sLinuxVMNameParts(vmName)
		if err != nil {
			return 0, err
		}
	} else if osType == compute.OperatingSystemTypesWindows {
		_, _, _, agentIndex, err = windowsVMNameParts(vmName)
		if err != nil {
			return 0, err
		}
	}

	return agentIndex, nil
}

// getLastSegment gets the last segment (splitting by '/'.)
func getLastSegment(ID string) (string, error) {
	parts := strings.Split(strings.TrimSpace(ID), "/")
	name := parts[len(parts)-1]
	if len(name) == 0 {
		return "", fmt.Errorf("identifier '/' not found in resource name %q", ID)
	}

	return name, nil
}

// readDeploymentParameters gets deployment parameters from paramFilePath.
func readDeploymentParameters(paramFilePath string) (map[string]interface{}, error) {
	contents, err := ioutil.ReadFile(paramFilePath)
	if err != nil {
		klog.Errorf("Failed to read deployment parameters from file %q: %v", paramFilePath, err)
		return nil, err
	}

	deploymentParameters := make(map[string]interface{})
	if err := json.Unmarshal(contents, &deploymentParameters); err != nil {
		klog.Errorf("Failed to unmarshal deployment parameters from file %q: %v", paramFilePath, err)
		return nil, err
	}

	if v, ok := deploymentParameters["parameters"]; ok {
		return v.(map[string]interface{}), nil
	}

	return nil, fmt.Errorf("failed to get deployment parameters from file %s", paramFilePath)
}

func getContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func getContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// checkExistsFromError inspects an error and returns a true if err is nil,
// false if error is an autorest.Error with StatusCode=404 and will return the
// error back if error is another status code or another type of error.
func checkResourceExistsFromError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	v, ok := err.(autorest.DetailedError)
	if !ok {
		return false, err
	}
	if v.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, v
}

func checkResourceExistsFromRetryError(err *retry.Error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if err.HTTPStatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, err.Error()
}

// isSuccessHTTPResponse determines if the response from an HTTP request suggests success
func isSuccessHTTPResponse(resp *http.Response, err error) (isSuccess bool, realError error) {
	if err != nil {
		return false, err
	}

	if resp != nil {
		// HTTP 2xx suggests a successful response
		if 199 < resp.StatusCode && resp.StatusCode < 300 {
			return true, nil
		}

		return false, fmt.Errorf("failed with HTTP status code %d", resp.StatusCode)
	}

	// This shouldn't happen, it only ensures all exceptions are handled.
	return false, fmt.Errorf("failed with unknown error")
}

// convertResourceGroupNameToLower converts the resource group name in the resource ID to be lowered.
func convertResourceGroupNameToLower(resourceID string) (string, error) {
	matches := azureResourceGroupNameRE.FindStringSubmatch(resourceID)
	if len(matches) != 2 {
		return "", fmt.Errorf("%q isn't in Azure resource ID format", resourceID)
	}

	resourceGroup := matches[1]
	return strings.Replace(resourceID, resourceGroup, strings.ToLower(resourceGroup), 1), nil
}

// isAzureRequestsThrottled returns true when the err is http.StatusTooManyRequests (429),
// and when err shows the requests was not executed due to an ongoing throttling period.
func isAzureRequestsThrottled(rerr *retry.Error) bool {
	klog.V(6).Infof("isAzureRequestsThrottled: starts for error %v", rerr)
	if rerr == nil {
		return false
	}

	if rerr.HTTPStatusCode == 0 && rerr.RetryAfter.After(time.Now()) {
		return true
	}

	return rerr.HTTPStatusCode == http.StatusTooManyRequests
}

func isRunningVmPowerState(powerState string) bool {
	return powerState == vmPowerStateRunning || powerState == vmPowerStateStarting
}

func isKnownVmPowerState(powerState string) bool {
	knownPowerStates := map[string]bool{
		vmPowerStateStarting:     true,
		vmPowerStateRunning:      true,
		vmPowerStateStopping:     true,
		vmPowerStateStopped:      true,
		vmPowerStateDeallocating: true,
		vmPowerStateDeallocated:  true,
		vmPowerStateUnknown:      true,
	}
	return knownPowerStates[powerState]
}

func vmPowerStateFromStatuses(statuses []compute.InstanceViewStatus) string {
	for _, status := range statuses {
		if status.Code == nil || !isKnownVmPowerState(*status.Code) {
			continue
		}
		return *status.Code
	}

	// PowerState is not set if the VM is still creating (or has failed creation)
	return vmPowerStateUnknown
}

// strconv.ParseInt, but for int
func parseInt32(s string, base int) (int, error) {
	val, err := strconv.ParseInt(s, base, 32)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

// strconv.ParseFloat, but for float32
func parseFloat32(s string) (float32, error) {
	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, err
	}
	return float32(val), nil
}
