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
type ReinstallServerWithCloudInitRequest struct {
	ServerId string                                   `json:"server_id"`
	Body     *ReinstallServerWithCloudInitRequestBody `json:"body,omitempty"`
}

func (o ReinstallServerWithCloudInitRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ReinstallServerWithCloudInitRequest", string(data)}, " ")
}
