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

// Request Object
type ListServerBlockDevicesRequest struct {
	ServerId string `json:"server_id"`
}

func (o ListServerBlockDevicesRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListServerBlockDevicesRequest", string(data)}, " ")
}
