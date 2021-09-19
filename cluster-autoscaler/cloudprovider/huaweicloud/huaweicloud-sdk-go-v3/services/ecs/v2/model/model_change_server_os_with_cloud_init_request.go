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
type ChangeServerOsWithCloudInitRequest struct {
	ServerId string                                  `json:"server_id"`
	Body     *ChangeServerOsWithCloudInitRequestBody `json:"body,omitempty"`
}

func (o ChangeServerOsWithCloudInitRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ChangeServerOsWithCloudInitRequest", string(data)}, " ")
}
