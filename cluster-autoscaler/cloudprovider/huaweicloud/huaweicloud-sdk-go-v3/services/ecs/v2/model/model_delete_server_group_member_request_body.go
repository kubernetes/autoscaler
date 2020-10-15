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

// This is a auto create Body Object
type DeleteServerGroupMemberRequestBody struct {
	RemoveMember *ServerGroupMember `json:"remove_member"`
}

func (o DeleteServerGroupMemberRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteServerGroupMemberRequestBody", string(data)}, " ")
}
