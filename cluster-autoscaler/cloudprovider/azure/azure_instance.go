/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"k8s.io/klog/v2"
)

// GetVMSSTypeStatically uses static list of vmss generated at azure_instance_types.go to fetch vmss instance information.
// It is declared as a variable for testing purpose.
var GetVMSSTypeStatically = func(template compute.VirtualMachineScaleSet) (*InstanceType, error) {
	var vmssType *InstanceType

	for k := range InstanceTypes {
		if strings.EqualFold(k, *template.Sku.Name) {
			vmssType = InstanceTypes[k]
			break
		}
	}

	promoRe := regexp.MustCompile(`(?i)_promo`)
	if promoRe.MatchString(*template.Sku.Name) {
		if vmssType == nil {
			// We didn't find an exact match but this is a promo type, check for matching standard
			klog.V(4).Infof("No exact match found for %s, checking standard types", *template.Sku.Name)
			skuName := promoRe.ReplaceAllString(*template.Sku.Name, "")
			for k := range InstanceTypes {
				if strings.EqualFold(k, skuName) {
					vmssType = InstanceTypes[k]
					break
				}
			}
		}
	}
	if vmssType == nil {
		return vmssType, fmt.Errorf("instance type %q not supported", *template.Sku.Name)
	}
	return vmssType, nil
}

// GetVMSSTypeDynamically fetched vmss instance information using sku api calls.
// It is declared as a variable for testing purpose.
var GetVMSSTypeDynamically = func(template compute.VirtualMachineScaleSet, azCache *azureCache) (InstanceType, error) {
	ctx := context.Background()
	var vmssType InstanceType

	sku, err := azCache.GetSKU(ctx, *template.Sku.Name, *template.Location)
	if err != nil {
		// We didn't find an exact match but this is a promo type, check for matching standard
		promoRe := regexp.MustCompile(`(?i)_promo`)
		skuName := promoRe.ReplaceAllString(*template.Sku.Name, "")
		if skuName != *template.Sku.Name {
			klog.V(1).Infof("No exact match found for %q, checking standard type %q. Error %v", *template.Sku.Name, skuName, err)
			sku, err = azCache.GetSKU(ctx, skuName, *template.Location)
		}
		if err != nil {
			return vmssType, fmt.Errorf("instance type %q not supported. Error %v", *template.Sku.Name, err)
		}
	}

	vmssType.VCPU, err = sku.VCPU()
	if err != nil {
		klog.V(1).Infof("Failed to parse vcpu from sku %q %v", *template.Sku.Name, err)
		return vmssType, err
	}
	gpu, err := getGpuFromSku(sku)
	if err != nil {
		klog.V(1).Infof("Failed to parse gpu from sku %q %v", *template.Sku.Name, err)
		return vmssType, err
	}
	vmssType.GPU = gpu

	memoryGb, err := sku.Memory()
	if err != nil {
		klog.V(1).Infof("Failed to parse memoryMb from sku %q %v", *template.Sku.Name, err)
		return vmssType, err
	}
	vmssType.MemoryMb = int64(memoryGb) * 1024

	return vmssType, nil
}
