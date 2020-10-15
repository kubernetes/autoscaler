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
type AddServerGroupMemberRequest struct {
	ServerGroupId string                           `json:"server_group_id"`
	Body          *AddServerGroupMemberRequestBody `json:"body,omitempty"`
}

func (o AddServerGroupMemberRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"AddServerGroupMemberRequest", string(data)}, " ")
}
