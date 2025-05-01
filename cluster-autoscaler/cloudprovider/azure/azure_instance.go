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

	"k8s.io/klog/v2"
)

// GetInstanceTypeStatically uses static list of vmss generated at azure_instance_types.go to fetch vmss instance information.
// It is declared as a variable for testing purpose.
var GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
	var instanceType *InstanceType

	for k := range InstanceTypes {
		if strings.EqualFold(k, template.SkuName) {
			instanceType = InstanceTypes[k]
			break
		}
	}

	promoRe := regexp.MustCompile(`(?i)_promo`)
	if promoRe.MatchString(template.SkuName) {
		if instanceType == nil {
			// We didn't find an exact match but this is a promo type, check for matching standard
			klog.V(4).Infof("No exact match found for %s, checking standard types", template.SkuName)
			skuName := promoRe.ReplaceAllString(template.SkuName, "")
			for k := range InstanceTypes {
				if strings.EqualFold(k, skuName) {
					instanceType = InstanceTypes[k]
					break
				}
			}
		}
	}
	if instanceType == nil {
		return instanceType, fmt.Errorf("instance type %q not supported", template.SkuName)
	}
	return instanceType, nil
}

// GetInstanceTypeDynamically fetched vmss instance information using sku api calls.
// It is declared as a variable for testing purpose.
var GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
	ctx := context.Background()
	var instanceType InstanceType

	sku, err := azCache.GetSKU(ctx, template.SkuName, template.Location)
	if err != nil {
		// We didn't find an exact match but this is a promo type, check for matching standard
		promoRe := regexp.MustCompile(`(?i)_promo`)
		skuName := promoRe.ReplaceAllString(template.SkuName, "")
		if skuName != template.SkuName {
			klog.V(1).Infof("No exact match found for %q, checking standard type %q. Error %v", template.SkuName, skuName, err)
			sku, err = azCache.GetSKU(ctx, skuName, template.Location)
		}
		if err != nil {
			return instanceType, fmt.Errorf("instance type %q not supported. Error %v", template.SkuName, err)
		}
	}

	instanceType.VCPU, err = sku.VCPU()
	if err != nil {
		klog.V(1).Infof("Failed to parse vcpu from sku %q %v", template.SkuName, err)
		return instanceType, err
	}
	gpu, err := getGpuFromSku(sku)
	if err != nil {
		klog.V(1).Infof("Failed to parse gpu from sku %q %v", template.SkuName, err)
		return instanceType, err
	}
	instanceType.GPU = gpu

	memoryGb, err := sku.Memory()
	if err != nil {
		klog.V(1).Infof("Failed to parse memoryMb from sku %q %v", template.SkuName, err)
		return instanceType, err
	}
	instanceType.MemoryMb = int64(memoryGb) * 1024

	return instanceType, nil
}
