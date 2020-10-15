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
type CreateServerGroupRequest struct {
	Body *CreateServerGroupRequestBody `json:"body,omitempty"`
}

func (o CreateServerGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateServerGroupRequest", string(data)}, " ")
}
