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
type DeleteServerMetadataRequest struct {
	Key      string `json:"key"`
	ServerId string `json:"server_id"`
}

func (o DeleteServerMetadataRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteServerMetadataRequest", string(data)}, " ")
}
