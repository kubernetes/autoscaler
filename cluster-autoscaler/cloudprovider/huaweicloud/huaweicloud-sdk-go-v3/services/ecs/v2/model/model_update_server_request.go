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
type UpdateServerRequest struct {
	ServerId string                   `json:"server_id"`
	Body     *UpdateServerRequestBody `json:"body,omitempty"`
}

func (o UpdateServerRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateServerRequest", string(data)}, " ")
}
