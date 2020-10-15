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
type UpdateServerMetadataRequest struct {
	ServerId string                           `json:"server_id"`
	Body     *UpdateServerMetadataRequestBody `json:"body,omitempty"`
}

func (o UpdateServerMetadataRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateServerMetadataRequest", string(data)}, " ")
}
