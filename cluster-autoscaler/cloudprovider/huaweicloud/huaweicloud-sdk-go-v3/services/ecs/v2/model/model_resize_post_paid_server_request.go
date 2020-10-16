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
type ResizePostPaidServerRequest struct {
	ServerId string                           `json:"server_id"`
	Body     *ResizePostPaidServerRequestBody `json:"body,omitempty"`
}

func (o ResizePostPaidServerRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ResizePostPaidServerRequest", string(data)}, " ")
}
