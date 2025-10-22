package proxmox

import (
	"fmt"
	"strings"

	goproxmox "github.com/luthermonson/go-proxmox"
	"k8s.io/apimachinery/pkg/api/resource"
)

func parseVMTags(tags string) map[string]string {
	result := make(map[string]string)
	if tags == "" {
		return result
	}

	for _, tag := range strings.Split(tags, ";") {
		if strings.Contains(tag, "node-group_") {
			parts := strings.SplitN(tag, "_", 2)
			result[parts[0]] = parts[1]
		} else {
			result[tag] = ""
		}
	}
	return result
}

func extractUUID(vmc *goproxmox.VirtualMachineConfig) string {
	var uuid string
	if vmc != nil && strings.Contains(vmc.SMBios1, "uuid=") {
		splits := strings.Split(vmc.SMBios1, ",")
		for _, split := range splits {
			if strings.HasPrefix(split, "uuid=") {
				uuid = strings.TrimPrefix(split, "uuid=")
			}
		}
	}

	return uuid
}

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", proxmoxProviderIDPrefix, nodeID)
}

// toNodeID returns a node or VM ID from the given provider ID.
func toNodeID(providerID string) string {
	if strings.HasPrefix(providerID, proxmoxProviderIDPrefix) {
		return strings.TrimPrefix(providerID, proxmoxProviderIDPrefix)
	}
	return providerID
}

// Helper function to create resource.Quantity from int64
func newIntQuantity(value int64) *resource.Quantity {
	q := resource.NewQuantity(value, resource.DecimalSI)
	return q
}
