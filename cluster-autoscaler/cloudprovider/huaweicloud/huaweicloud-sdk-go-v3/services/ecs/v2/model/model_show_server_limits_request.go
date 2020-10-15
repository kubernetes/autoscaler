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
type ShowServerLimitsRequest struct {
}

func (o ShowServerLimitsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerLimitsRequest", string(data)}, " ")
}
