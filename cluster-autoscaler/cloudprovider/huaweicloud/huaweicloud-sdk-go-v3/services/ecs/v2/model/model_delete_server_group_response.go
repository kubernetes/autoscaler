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
type DeleteServerGroupResponse struct {
}

func (o DeleteServerGroupResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteServerGroupResponse", string(data)}, " ")
}
