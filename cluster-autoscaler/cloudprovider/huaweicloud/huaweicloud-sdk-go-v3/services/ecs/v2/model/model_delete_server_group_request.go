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
type DeleteServerGroupRequest struct {
	ServerGroupId string `json:"server_group_id"`
}

func (o DeleteServerGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteServerGroupRequest", string(data)}, " ")
}
