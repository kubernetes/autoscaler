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
type NovaShowServerResponse struct {
	Server *NovaServer `json:"server,omitempty"`
}

func (o NovaShowServerResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaShowServerResponse", string(data)}, " ")
}
