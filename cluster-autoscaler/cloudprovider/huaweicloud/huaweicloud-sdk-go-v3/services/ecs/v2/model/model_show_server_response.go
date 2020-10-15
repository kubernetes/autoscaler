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
type ShowServerResponse struct {
	Server *ServerDetail `json:"server,omitempty"`
}

func (o ShowServerResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerResponse", string(data)}, " ")
}
