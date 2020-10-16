/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type NovaListAvailabilityZonesResponse struct {
	// 可用域信息。
	AvailabilityZoneInfo *[]NovaAvailabilityZone `json:"availabilityZoneInfo,omitempty"`
}

func (o NovaListAvailabilityZonesResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListAvailabilityZonesResponse", string(data)}, " ")
}
