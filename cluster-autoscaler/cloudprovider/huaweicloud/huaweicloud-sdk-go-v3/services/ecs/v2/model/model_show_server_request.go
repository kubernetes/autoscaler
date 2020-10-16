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
type ShowServerRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowServerRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerRequest", string(data)}, " ")
}
