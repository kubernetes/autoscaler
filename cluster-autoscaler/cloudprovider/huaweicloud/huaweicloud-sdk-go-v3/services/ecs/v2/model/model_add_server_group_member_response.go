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
type AddServerGroupMemberResponse struct {
}

func (o AddServerGroupMemberResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"AddServerGroupMemberResponse", string(data)}, " ")
}
