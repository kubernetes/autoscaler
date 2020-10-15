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
type ResetServerPasswordRequest struct {
	ServerId string                          `json:"server_id"`
	Body     *ResetServerPasswordRequestBody `json:"body,omitempty"`
}

func (o ResetServerPasswordRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ResetServerPasswordRequest", string(data)}, " ")
}
