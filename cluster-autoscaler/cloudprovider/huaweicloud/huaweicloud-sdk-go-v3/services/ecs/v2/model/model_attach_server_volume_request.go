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
type AttachServerVolumeRequest struct {
	ServerId string                         `json:"server_id"`
	Body     *AttachServerVolumeRequestBody `json:"body,omitempty"`
}

func (o AttachServerVolumeRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"AttachServerVolumeRequest", string(data)}, " ")
}
