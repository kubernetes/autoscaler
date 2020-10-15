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
type ShowServerRemoteConsoleRequest struct {
	ServerId string                              `json:"server_id"`
	Body     *ShowServerRemoteConsoleRequestBody `json:"body,omitempty"`
}

func (o ShowServerRemoteConsoleRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerRemoteConsoleRequest", string(data)}, " ")
}
