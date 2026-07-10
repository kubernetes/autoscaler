package skewer

import "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"

// Wrap takes an array of compute resource skus and wraps them into an
// array of our richer type.
func Wrap(in []*armcompute.ResourceSKU) []SKU {
	out := make([]SKU, len(in))
	for index, value := range in {
		if value != nil {
			out[index] = SKU(*value)
		}
	}
	return out
}
