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
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/golang/glog"
	"golang.org/x/crypto/pkcs12"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/pkg/version"
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

	// resource ids
	nsgID = "nsgID"
	rtID  = "routeTableID"

	k8sLinuxVMNamingFormat         = "^[0-9a-zA-Z]{3}-(.+)-([0-9a-fA-F]{8})-{0,2}([0-9]+)$"
	k8sLinuxVMAgentPoolNameIndex   = 1
	k8sLinuxVMAgentClusterIDIndex  = 2
	k8sLinuxVMAgentIndexArrayIndex = 3

	k8sWindowsVMNamingFormat               = "^([a-fA-F0-9]{5})([0-9a-zA-Z]{3})([a-zA-Z0-9]{4,6})$"
	k8sWindowsVMAgentPoolPrefixIndex       = 1
	k8sWindowsVMAgentOrchestratorNameIndex = 2
	k8sWindowsVMAgentPoolInfoIndex         = 3
)

var (
	vmnameLinuxRegexp   = regexp.MustCompile(k8sLinuxVMNamingFormat)
	vmnameWindowsRegexp = regexp.MustCompile(k8sWindowsVMNamingFormat)
)

// decodePkcs12 decodes a PKCS#12 client certificate by extracting the public certificate and
// the private RSA key
func decodePkcs12(pkcs []byte, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, certificate, err := pkcs12.Decode(pkcs, password)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding the PKCS#12 client certificate: %v", err)
	}
	rsaPrivateKey, isRsaKey := privateKey.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil, nil, fmt.Errorf("PKCS#12 certificate must contain a RSA private key")
	}

	return certificate, rsaPrivateKey, nil
}

// configureUserAgent configures the autorest client with a user agent that
// includes "autoscaler" and the full client version string
// example:
// Azure-SDK-for-Go/7.0.1-beta arm-network/2016-09-01; cluster-autoscaler/v1.7.0-alpha.2.711+a2fadef8170bb0-dirty;
func configureUserAgent(client *autorest.Client) {
	k8sVersion := version.Get().GitVersion
	client.UserAgent = fmt.Sprintf("%s; cluster-autoscaler/%s", client.UserAgent, k8sVersion)
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
			glog.Warningf("Template improperly formatted for resource")
			continue
		}

		resourceType, ok := resourceMap[typeFieldName].(string)
		if ok && resourceType == nsgResourceType {
			if nsgIndex != -1 {
				err := fmt.Errorf("Found 2 resources with type %s in the template. There should only be 1", nsgResourceType)
				glog.Errorf(err.Error())
				return err
			}
			nsgIndex = index
		}
		if ok && resourceType == rtResourceType {
			if rtIndex != -1 {
				err := fmt.Errorf("Found 2 resources with type %s in the template. There should only be 1", rtResourceType)
				glog.Warningf(err.Error())
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
		err := fmt.Errorf("Found no resources with type %s in the template. There should have been 1", nsgResourceType)
		glog.Errorf(err.Error())
		return err
	}
	if rtIndex == -1 {
		glog.Infof("Found no resources with type %s in the template.", rtResourceType)
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
			glog.Warningf("Template improperly formatted")
			continue
		}

		resourceType, ok := resourceMap[typeFieldName].(string)
		if !ok || resourceType != vmResourceType {
			resourceName, ok := resourceMap[nameFieldName].(string)
			if !ok {
				glog.Warningf("Template improperly formatted")
				continue
			}
			if strings.Contains(resourceName, "variables('masterVMNamePrefix')") && resourceType == vmExtensionType {
				indexesToRemove = append(indexesToRemove, index)
			}
			continue
		}

		resourceName, ok := resourceMap[nameFieldName].(string)
		if !ok {
			glog.Warningf("Template improperly formatted")
			continue
		}

		// make sure this is only modifying the master vms
		if !strings.Contains(resourceName, "variables('masterVMNamePrefix')") {
			continue
		}

		resourceProperties, ok := resourceMap[propertiesFieldName].(map[string]interface{})
		if !ok {
			glog.Warningf("Template improperly formatted")
			continue
		}

		hardwareProfile, ok := resourceProperties[hardwareProfileFieldName].(map[string]interface{})
		if !ok {
			glog.Warningf("Template improperly formatted")
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
		glog.Warningf("Template improperly formatted")
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
		glog.Warningf("Template improperly formatted. Could not find: %s", storageProfileFieldName)
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
		return "", "", -1, fmt.Errorf("Error parsing VM Name: %v", err)
	}

	return vmNameParts[k8sLinuxVMAgentPoolNameIndex], vmNameParts[k8sLinuxVMAgentClusterIDIndex], vmNum, nil
}

// windowsVMNameParts returns parts of Windows VM name e.g: 50621k8s9000
func windowsVMNameParts(vmName string) (poolPrefix string, acsStr string, poolIndex int, agentIndex int, err error) {
	vmNameParts := vmnameWindowsRegexp.FindStringSubmatch(vmName)
	if len(vmNameParts) != 4 {
		return "", "", -1, -1, fmt.Errorf("resource name was missing from identifier")
	}

	poolPrefix = vmNameParts[k8sWindowsVMAgentPoolPrefixIndex]
	acsStr = vmNameParts[k8sWindowsVMAgentOrchestratorNameIndex]
	poolInfo := vmNameParts[k8sWindowsVMAgentPoolInfoIndex]

	poolIndex, err = strconv.Atoi(poolInfo[:3])
	if err != nil {
		return "", "", -1, -1, fmt.Errorf("Error parsing VM Name: %v", err)
	}

	agentIndex, err = strconv.Atoi(poolInfo[3:])
	if err != nil {
		return "", "", -1, -1, fmt.Errorf("Error parsing VM Name: %v", err)
	}

	return poolPrefix, acsStr, poolIndex, agentIndex, nil
}

// GetVMNameIndex return the index of VM in the node pools.
func GetVMNameIndex(osType compute.OperatingSystemTypes, vmName string) (int, error) {
	var agentIndex int
	var err error
	if osType == compute.Linux {
		_, _, agentIndex, err = k8sLinuxVMNameParts(vmName)
		if err != nil {
			return 0, err
		}
	} else if osType == compute.Windows {
		_, _, _, agentIndex, err = windowsVMNameParts(vmName)
		if err != nil {
			return 0, err
		}
	}

	return agentIndex, nil
}

func matchDiscoveryConfig(labels map[string]*string, configs []cloudprovider.LabelAutoDiscoveryConfig) bool {
	if len(configs) == 0 {
		return false
	}

	for _, c := range configs {
		if len(c.Selector) == 0 {
			return false
		}

		for k, v := range c.Selector {
			value, ok := labels[k]
			if !ok {
				return false
			}

			if len(v) > 0 {
				if value == nil || *value != v {
					return false
				}
			}
		}
	}

	return true
}

func validateConfig(cfg *Config) error {
	if cfg.ResourceGroup == "" {
		return fmt.Errorf("resource group not set")
	}

	if cfg.SubscriptionID == "" {
		return fmt.Errorf("subscription ID not set")
	}

	if cfg.TenantID == "" {
		return fmt.Errorf("tenant ID not set")
	}

	if cfg.AADClientID == "" {
		return fmt.Errorf("ARM Client ID not set")
	}

	if cfg.VMType == vmTypeStandard {
		if cfg.Deployment == "" {
			return fmt.Errorf("deployment not set")
		}

		if len(cfg.DeploymentParameters) == 0 {
			return fmt.Errorf("deploymentParameters not set")
		}
	}

	return nil
}

// getLastSegment gets the last segment splited by '/'.
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
		glog.Errorf("Failed to read deployment parameters from file %q: %v", paramFilePath, err)
		return nil, err
	}

	deploymentParameters := make(map[string]interface{})
	if err := json.Unmarshal(contents, &deploymentParameters); err != nil {
		glog.Errorf("Failed to unmarshal deployment parameters from file %q: %v", paramFilePath, err)
		return nil, err
	}

	if v, ok := deploymentParameters["parameters"]; ok {
		return v.(map[string]interface{}), nil
	}

	return nil, fmt.Errorf("failed to get deployment parameters from file %s", paramFilePath)
}
