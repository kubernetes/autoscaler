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
	"strings"

	"github.com/Azure/skewer"
	"github.com/pkg/errors"
)

var (
	// NvidiaEnabledSKUs represents a list of NVIDIA gpus.
	// If a new GPU sku becomes available, add a key to this map, but only if you have a confirmation
	//   that we have an agreement with NVIDIA for this specific gpu.
	NvidiaEnabledSKUs = map[string]bool{
		// K80
		"standard_nc6":   true,
		"standard_nc12":  true,
		"standard_nc24":  true,
		"standard_nc24r": true,
		// M60
		"standard_nv6":      true,
		"standard_nv12":     true,
		"standard_nv12s_v3": true,
		"standard_nv24":     true,
		"standard_nv24s_v3": true,
		"standard_nv24r":    true,
		"standard_nv48s_v3": true,
		// P40
		"standard_nd6s":   true,
		"standard_nd12s":  true,
		"standard_nd24s":  true,
		"standard_nd24rs": true,
		// P100
		"standard_nc6s_v2":   true,
		"standard_nc12s_v2":  true,
		"standard_nc24s_v2":  true,
		"standard_nc24rs_v2": true,
		// V100
		"standard_nc6s_v3":   true,
		"standard_nc12s_v3":  true,
		"standard_nc24s_v3":  true,
		"standard_nc24rs_v3": true,
		"standard_nd40s_v3":  true,
		"standard_nd40rs_v2": true,
		// T4
		"standard_nc4as_t4_v3":  true,
		"standard_nc8as_t4_v3":  true,
		"standard_nc16as_t4_v3": true,
		"standard_nc64as_t4_v3": true,
		// A100 40GB
		"standard_nd96asr_v4":       true,
		"standard_nd112asr_a100_v4": true,
		"standard_nd120asr_a100_v4": true,
		// A100 80GB
		"standard_nd96amsr_a100_v4":  true,
		"standard_nd112amsr_a100_v4": true,
		"standard_nd120amsr_a100_v4": true,
		// A100 PCIE 80GB
		"standard_nc24ads_a100_v4": true,
		"standard_nc48ads_a100_v4": true,
		"standard_nc96ads_a100_v4": true,
	}
)

// isNvidiaEnabledSKU determines if an VM SKU has nvidia driver support.
func isNvidiaEnabledSKU(vmSize string) bool {
	// Trim the optional _Promo suffix.
	vmSize = strings.ToLower(vmSize)
	vmSize = strings.TrimSuffix(vmSize, "_promo")
	return NvidiaEnabledSKUs[vmSize]
}

// getGpuFromSku extracts gpu information from vmss sku.
func getGpuFromSku(sku skewer.SKU) (int64, error) {
	errCapabilityValueNil := &skewer.ErrCapabilityValueNil{}
	errCapabilityNotFound := &skewer.ErrCapabilityNotFound{}

	var value int64
	value, err := sku.GetCapabilityIntegerQuantity("GPUs")
	if err != nil {
		// In case of an error, SKU api returns -1 as the value.
		// Updating value=0 if it's a non-gpu sku or value is nil.
		if errors.As(err, &errCapabilityValueNil) || errors.As(err, &errCapabilityNotFound) {
			value = 0
		} else {
			return value, err
		}
	}
	return value, nil
}
