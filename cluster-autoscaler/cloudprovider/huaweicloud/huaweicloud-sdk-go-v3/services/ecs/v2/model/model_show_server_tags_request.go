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
type ShowServerTagsRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowServerTagsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowServerTagsRequest", string(data)}, " ")
}
