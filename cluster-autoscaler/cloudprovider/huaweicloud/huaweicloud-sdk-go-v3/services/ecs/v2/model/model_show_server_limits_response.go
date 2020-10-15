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
type ShowServerLimitsResponse struct {
	Absolute *ServerLimits `json:"absolute,omitempty"`
}

func (o ShowServerLimitsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerLimitsResponse", string(data)}, " ")
}
