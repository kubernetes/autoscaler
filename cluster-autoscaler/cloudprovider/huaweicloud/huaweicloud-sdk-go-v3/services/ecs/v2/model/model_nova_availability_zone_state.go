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

// 可用域的状态
type NovaAvailabilityZoneState struct {
	// 可用域状态。
	Available bool `json:"available"`
}

func (o NovaAvailabilityZoneState) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaAvailabilityZoneState", string(data)}, " ")
}
